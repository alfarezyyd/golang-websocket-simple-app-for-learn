package main

import (
	"github.com/gorilla/websocket"
	"golang-websocket-simple-app-for-learn/helper"
	"log"
	"sync"
	"time"
)

func main() {
	c, _, err := websocket.DefaultDialer.Dial("ws://localhost:3000/request", nil)
	waitGroup := &sync.WaitGroup{}

	helper.LogIfError(err)
	defer waitGroup.Done()
	defer c.Close()

	c.SetPingHandler(func(appData string) error {
		log.Printf("Received Ping : %s", appData)
		err = c.WriteControl(websocket.PongMessage, []byte{}, time.Time{})
		helper.LogIfError(err)
		return nil
	})
	waitGroup.Add(1)
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println(err)
				break
			}
			log.Printf("Receive : %s", message)
		}
	}()

	newTicker := time.NewTicker(2 * time.Second)
	defer newTicker.Stop()

	waitGroup.Wait()

}
