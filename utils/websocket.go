package utils

import (
	"fmt"
	"net/http"
	"os"

	"github.com/delta/codecharacter-lsp-2023/controllers"
	"github.com/delta/codecharacter-lsp-2023/models"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{}

func InitWebsocket(c echo.Context) error {
	var ws models.WebsocketConnection
	id := uuid.New()
	ws.ID = id
	language := c.Param("language")
	if language != "cpp" && language != "java" && language != "python" {
		return c.String(http.StatusBadRequest, "Invalid Language")
	}
	ws.Language = language
	fmt.Println("WS Connection Created with ID : ", id, " and Language : ", language)
	wsConn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return c.String(http.StatusBadRequest, "Error Upgrading to Websocket Connection")
	}
	ws.Connection = wsConn
	err = createWebsocketConnection(&ws)
	if err != nil {
		return c.String(http.StatusBadGateway, "Something went wrong, contact the event administrator.")
	}
	return nil
}

func dropConnection(ws *models.WebsocketConnection) {
	err := os.RemoveAll(ws.WorkspacePath)
	if err != nil {
		fmt.Println(err)
	}
	if ws.LSPServer != nil {
		err = ws.LSPServer.Process.Signal(os.Interrupt)
		if err != nil {
			fmt.Println(err)
		}
		// Reads process exit state to remove the <defunct> process from the system process table
		err = ws.LSPServer.Wait()
		if err != nil {
			fmt.Println(err)
		}
	}
	ws.Connection.Close()
	fmt.Println("Dropped WS Connection : ", ws.ID)
}

func createWebsocketConnection(ws *models.WebsocketConnection) error {
	defer dropConnection(ws)

	ws.WorkspacePath = "workspaces/" + ws.ID.String()
	err := os.Mkdir(ws.WorkspacePath, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = CreateLSPServer(ws)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = listen(ws)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func listen(ws *models.WebsocketConnection) error {
	for {
		fmt.Println("Listening for Messages")
		_, messageBytes, err := ws.Connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Println("WS Connection ", ws.ID, " closing with error : ", err)
				return err
			}
			return nil
		}
		err = controllers.HandleMessage(ws, messageBytes)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
}