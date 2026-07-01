package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)
var upgrader = websocket.Upgrader{
	CheckOrigin:  func(r *http.Request) bool {return true},
}
func wsHandler (w http.ResponseWriter, r *http.Request){
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil{
		log.Fatal(err)
	}

	log.Println("client connected")
	for {
		msgType, msg, err := conn.ReadMessage()d
		if err != nil{
			log.Fatal(err)
			break
		}

		log.Printf("Recieved: %s",msg)

		err2 := conn.WriteMessage(msgType, msg)
		if err2 != nil {
					log.Println("write error:", err2)
					break
				}

	}

	log.Println("client disconnected")
}
func pingHandler (w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "pinged!"})
}

func main(){
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", pingHandler)
	mux.HandleFunc("/ws", wsHandler)
	log.Println("Server starte on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal("ERROR: ", err)
	}

}
