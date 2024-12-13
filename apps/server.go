package apps

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"ij4l.github.com/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Server struct {
	Counter       int
	CounterChange chan int
	CounterMu     sync.Mutex
	Connections   []*websocket.Conn
	ConnMu        sync.Mutex
}

func (s *Server) Web(ctx *gin.Context) {
	ctx.File("./static/index.html")
}

func (s *Server) WatchCounter() {
	for value := range s.CounterChange {
		fmt.Println("Counter value:", value)
		s.broadcastData(value)
	}
}

func (s *Server) broadcastData(newValue int) {
	s.ConnMu.Lock()
	defer s.ConnMu.Unlock()

	for _, conn := range s.Connections {
		if err := conn.WriteJSON(map[string]int{"data": newValue}); err != nil {
			log.Println("Error broadcasting to client:", err)
		}
	}
}

func (s *Server) updateData() {
	s.CounterMu.Lock()
	defer s.CounterMu.Unlock()

	data := s.Counter
	s.CounterChange <- data
}

func (s *Server) Increment(ctx *gin.Context) {
	s.CounterMu.Lock()
	s.Counter++
	currentCount := s.Counter
	s.CounterMu.Unlock()

	s.updateData()
	ctx.JSON(http.StatusOK, gin.H{"count": currentCount})
}

func (s *Server) Decrement(ctx *gin.Context) {
	s.CounterMu.Lock()
	s.Counter--
	currentCount := s.Counter
	s.CounterMu.Unlock()

	s.updateData()
	ctx.JSON(http.StatusOK, gin.H{"count": currentCount})
}

func (s *Server) WsHandler(ctx *gin.Context) {
	conn, err := utils.Upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("webSocket upgrade failed: %s", err)})
		return
	}
	defer conn.Close()

	s.ConnMu.Lock()
	s.Connections = append(s.Connections, conn)
	s.ConnMu.Unlock()

	defer func() {
		s.ConnMu.Lock()
		for i, c := range s.Connections {
			if c == conn {
				s.Connections = append(s.Connections[:i], s.Connections[i+1:]...)
				break
			}
		}
		s.ConnMu.Unlock()
	}()

	s.CounterMu.Lock()
	currentCount := s.Counter
	s.CounterMu.Unlock()

	if err := conn.WriteJSON(map[string]int{"count": currentCount}); err != nil {
		fmt.Printf("failed to send initial count to client: %s\n", err)
	}
	select {}
}
