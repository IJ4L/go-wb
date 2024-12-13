package main

import (
	"ij4l.github.com/apps"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func main() {
	r := gin.Default()

	server := apps.Server{
		Counter:       0,
		CounterChange: make(chan int),
		Connections:   make([]*websocket.Conn, 0),
	}

	go server.WatchCounter()
	r.GET("/", server.Web)
	r.GET("/inc", server.Increment)
	r.GET("/dec", server.Decrement)
	r.GET("/ws", server.WsHandler)
	r.Run(":8080")
}
