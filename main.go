package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)
var upgrader = websocket.Upgrader{

	CheckOrigin:  func(r *http.Request) bool {return true},
}
func wsHandler (hub  *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request){
		conn, err := upgrader.Upgrade(w, r, nil)

		if err != nil{
			log.Println("upgrade error: ",err)
			return
		}

		hub.register <- conn

		defer func(){
			hub.unregister <- conn
		}()

		for{
			_, msg, err := conn.ReadMessage()
			if err != nil{
				log.Println("read error: ",err)
				return
			}
			hub.broadcast <- msg

		}
	}
}


func pingHandler (w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "pinged!"})
}

func main(){
	godotenv.Load()
	hub := NewHub()
	go hub.run()
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", pingHandler)
	mux.HandleFunc("/ws", wsHandler(hub))
	mux.HandleFunc("GET /debug", func(w http.ResponseWriter, r *http.Request){
		hub.mu.Lock()
		defer hub.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"connected_clients": len(hub.clients),
		})
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Server started on :" + port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal("ERROR: ", err)
	}
}
