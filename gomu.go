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
	http.Handle("/", http.FileServer(http.Dir("client")))

	http.Handle("/ws", websocket.Handler(func(ws *websocket.Conn) {
		network.Connection(eng, ws)
	}))

	http.ListenAndServe(":8080", nil)
}
