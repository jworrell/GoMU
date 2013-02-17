package main

import (
	"code.google.com/p/go.net/websocket"
	"github.com/jworrell/GoMU/engine"
	"github.com/jworrell/GoMU/network"
	"log"
	"net/http"
)

func main() {
	eng, err := engine.Init("data/world.json")
	if err != nil {
		log.Fatal(err)
	}

	go network.SocketServer(eng)

	http.Handle("/", http.FileServer(http.Dir("client")))

	http.Handle("/ws", websocket.Handler(func(ws *websocket.Conn) {
		network.WSConnection(eng, ws)
	}))

	http.ListenAndServe(":8080", nil)
}
