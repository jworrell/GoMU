package network

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"github.com/jworrell/GoMU/engine"
	"github.com/jworrell/GoMU/message"
	"github.com/jworrell/GoMU/object"
	"io"
	"log"
)

const WRITE_QUEUE = 16

func Connection(eng *engine.Engine, ws *websocket.Conn) {
	decoder := json.NewDecoder(ws)
	writer := make(chan *message.Message, WRITE_QUEUE)
	kill := make(chan bool)

	associatedObj := object.NewObject(nil, object.DUMMY_ID)
	associatedObj.SetWriter(writer)

	defer func(o *object.Object) {
		associatedObj.SetWriter(nil)
		kill <- true
	}(associatedObj)

	go func() {
		encoder := json.NewEncoder(ws)

		for {
			select {
			case writerMsg := <-writer:
				err := encoder.Encode(writerMsg)
				if err != nil {
					log.Println("Unexpected error (encode JSON): ", err)
				}

			case <-kill:
				return
			}

		}
	}()

	for {
		readerMsg := &message.Message{}
		err := decoder.Decode(readerMsg)
		if err != nil {
			if err != io.EOF {
				log.Println("Unexpected error (decode JSON): ", err)
			}
			return
		}

		readerMsg.Finalize()
		eng.Do(&associatedObj, readerMsg)
	}
}
