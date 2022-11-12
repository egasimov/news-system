package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	memphis_client "news-presenter/infrastructure"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func readerWithConsume(conn *websocket.Conn) {
	// read the incoming message
	messageType, p, errRead := conn.ReadMessage()
	if errRead != nil {
		log.Println(errRead)
		return
	}
	// print out that message for clarity to see what client sent
	log.Printf("[INFO] Client sent to server: %s", string(p))

	clientReqHandler := func(data any) {
		jsonData, err := json.Marshal(data)
		if err != nil {
			panic(err)
		}
		if err := conn.WriteMessage(messageType, jsonData); err != nil {
			log.Fatalln(err)
		}
	}

	//problem raised, nats : timeout when trying to pull message for the second time after given pullInterval
	errConsume := memphis_client.ConsumeFromMemphis(clientReqHandler)
	if errConsume != nil {
		log.Fatalln(errConsume)
	}
}

func readerWithFetch(conn *websocket.Conn) {
	for {
		select {
		case <-time.After(5 * time.Second):
			// read in a message
			messageType, p, errRead := conn.ReadMessage()
			if errRead != nil {
				log.Println(errRead)
				return
			}
			// print out that message for clarity
			log.Printf("[INFO] Client sent to server: %s", string(p))

			data, errConsume := memphis_client.FetchFromMemphis()
			if errConsume != nil {
				log.Fatalln(errConsume)
			}
			jsonData, err := json.Marshal(data)
			if err != nil {
				panic(err)
			}

			if err := conn.WriteMessage(messageType, jsonData); err != nil {
				log.Fatalln(err)
			}
		}
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home Page")
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	// upgrade this connection to a WebSocket
	// connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println("Client Connected")
	err = ws.WriteMessage(1, []byte("Hi Client!"))
	if err != nil {
		log.Println(err)
	}
	// listen indefinitely for new messages coming
	// through on our WebSocket connection
	readerWithConsume(ws)
}

func main() {
	//errLoad := godotenv.Load(".env.local")
	//if errLoad != nil {
	//	log.Fatalln(errLoad)
	//}

	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", wsEndpoint)

	log.Fatalln(http.ListenAndServe(":8081", nil))
}
