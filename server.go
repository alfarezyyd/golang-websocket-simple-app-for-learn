package main

import (
	"github.com/gorilla/websocket"
	"golang-websocket-simple-app-for-learn/helper"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	pongWait      = 10 * time.Second
	writeDeadline = 5 * time.Second
)

func upgradingConnection() http.HandlerFunc {
	upgradeConn := websocket.Upgrader{}
	return func(writer http.ResponseWriter, request *http.Request) {
		c, err := upgradeConn.Upgrade(writer, request, nil)
		defer c.Close()
		helper.LogIfError(err)
		waitGroup := &sync.WaitGroup{}
		waitGroup.Add(2)
		go senderFunc(c, waitGroup)
		go receiverFunc(c, waitGroup)
		waitGroup.Wait()
		log.Println("Websocket Handler is Done")
	}
}

func senderFunc(ws *websocket.Conn, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	messageTicker := time.NewTicker(1 * time.Second)
	defer messageTicker.Stop()

	pingTicker := time.NewTicker(5 * time.Second)
	defer pingTicker.Stop()
	counterVal := 0
breakLoop:
	for {
		select {
		case tickerTime := <-messageTicker.C:
			err := ws.SetWriteDeadline(time.Now().Add(writeDeadline))
			helper.LogIfError(err)
			err = ws.WriteMessage(websocket.TextMessage, []byte(tickerTime.String()))
			helper.LogIfError(err)
		case <-pingTicker.C:
			err := ws.SetWriteDeadline(time.Now().Add(writeDeadline))
			helper.LogIfError(err)
			err = ws.WriteControl(websocket.PingMessage, []byte("Ping Message"), time.Time{})
			helper.LogIfError(err)
		}

		if counterVal > 30 {
			break breakLoop
		}
		counterVal++
	}

	err := ws.SetWriteDeadline(time.Now().Add(writeDeadline))
	helper.LogIfError(err)
	err = ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	helper.LogIfError(err)
}

func receiverFunc(ws *websocket.Conn, waitGroup *sync.WaitGroup) {
	defer ws.Close()
	defer waitGroup.Done()

	err := ws.SetReadDeadline(time.Now().Add(pongWait))
	helper.LogIfError(err)
	ws.SetPongHandler(func(string) error {
		err = ws.SetReadDeadline(time.Now().Add(pongWait))
		helper.LogIfError(err)
		log.Println("Received Pong")
		return nil
	})

	for {
		_, messageData, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
		log.Println(messageData)
	}

}

func main() {
	newMux := http.NewServeMux()
	newMux.Handle("/request", upgradingConnection())
	newServer := http.Server{
		Addr:    "localhost:3000",
		Handler: newMux,
	}
	err := newServer.ListenAndServe()
	helper.LogIfError(err)
}
