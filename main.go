package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
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

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				sentry.CurrentHub().Recover(err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func main(){
	godotenv.Load()

	if err := sentry.Init(sentry.ClientOptions{
		Dsn: os.Getenv("SENTRY_DSN"),
	}); err != nil {
		log.Printf("sentry init error: %v", err)
	}
	defer sentry.Flush(2 * time.Second)

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
	if err := http.ListenAndServe(":"+port, recoveryMiddleware(mux)); err != nil {
		sentry.CaptureException(err)
		log.Fatal("ERROR: ", err)
	}
}
