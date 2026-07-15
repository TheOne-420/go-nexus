package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type Hub struct{
	clients 	map[*websocket.Conn]bool
	broadcast 	chan []byte
	register 	chan *websocket.Conn
	unregister 	chan *websocket.Conn
	mu          sync.Mutex
	rdb 		*redis.Client
}


func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func NewHub() *Hub{
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379")
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		opts = &redis.Options{
			Addr: getEnv("REDIS_ADDR", "localhost:6379"),
		}
	}
	rdb := redis.NewClient(opts)
	return &Hub{
		clients: make(map [*websocket.Conn]bool),
		broadcast: make(chan []byte),
		register: make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		rdb:        rdb,
	}
}
func (h *Hub) subscribeRedis (){
	sub := h.rdb.Subscribe(context.Background(), getEnv("REDIS_CHANNEL", "go-nexus"))
	ch := sub.Channel()

	for msg := range ch{
		payload := []byte(msg.Payload)

		h.mu.Lock()
		for conn:= range h.clients{
			if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil{
				fmt.Println("Write error: ", err);
				delete(h.clients, conn)
				conn.Close()
			}
		}
		h.mu.Unlock()
	}

}
func (h *Hub) run() {
	go h.subscribeRedis()
	for {
		select {
			case conn := <- h.register:
				h.mu.Lock()
				h.clients[conn]=true
				h.mu.Unlock()
				log.Println("A client registered")
				log.Println("Total clients:", len(h.clients))

			case conn := <- h.unregister:
				h.mu.Lock()
			 	delete(h.clients, conn)
				conn.Close()
				h.mu.Unlock()
				log.Println("A client unregistered")
				log.Println("Total clients:", len(h.clients))
			case msg := <- h.broadcast:
				err := h.rdb.Publish(context.Background(), getEnv("REDIS_CHANNEL", "go-nexus"), msg).Err()
				if err != nil{
					log.Println("redis publish error: ", err)
				}
		}
	}
}
