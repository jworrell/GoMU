package network

import (
	"github.com/jworrell/GoMU/engine"
	"github.com/jworrell/GoMU/message"
	"github.com/jworrell/GoMU/object"
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"log"
)

const WRITE_QUEUE = 32

func Connection(eng *engine.Engine, ws *websocket.Conn) {
	decoder := json.NewDecoder(ws)
	writer := make(chan *message.Message, WRITE_QUEUE)

	associatedObj := object.NewObject(object.DUMMY_ID)
	associatedObj.SetWriter(writer)

	go func() {
		encoder := json.NewEncoder(ws)

		for {
			writerMsg := <-writer
			err := encoder.Encode(writerMsg)

			if err != nil {
				log.Println("Error: ", err)
				break
			}
		}
	}()

	for {
		readerMsg := &message.Message{}
		err := decoder.Decode(readerMsg)

		if err != nil {
			log.Println("Error: ", err)
			break
		}

		eng.Do(&associatedObj, readerMsg)
	}
}
