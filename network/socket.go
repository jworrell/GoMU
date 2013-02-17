package network

import (
	"bufio"
	"github.com/jworrell/GoMU/engine"
	"github.com/jworrell/GoMU/message"
	"github.com/jworrell/GoMU/object"
	"log"
	"net"
	"strings"
)

const (
	BUFFER_SIZE = 4096
)

func SocketServer(eng *engine.Engine) {
	listener, err := net.Listen("tcp", ":9999")
	defer listener.Close()

	if err != nil {
		log.Printf("Error creating listener: %s\n", err)
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error:", err)
			return
		}

		go HandleClientSocket(eng, conn)
	}
}

func HandleClientSocket(eng *engine.Engine, conn net.Conn) {
	reader := bufio.NewReader(conn)
	writer := make(chan *message.Message, WRITE_QUEUE)
	kill := make(chan bool)

	associatedObj := object.NewObject(nil, object.DUMMY_ID)
	associatedObj.SetWriter(writer)

	defer func() {
		conn.Close()
		associatedObj.SetWriter(nil)
		kill <- true
	}()

	go func() {
		socketWriter := bufio.NewWriter(conn)

		for {
			select {
			case writerMsg := <-writer:
				_, err := socketWriter.WriteString(writerMsg.Data + "\n")

				if err != nil {
					log.Println("Unexpected error (write socket): ", err)
				}

				err = socketWriter.Flush()

				if err != nil {
					log.Println("Unexpected error (flush socket): ", err)
				}

			case <-kill:
				return
			}

		}
	}()

	for {
		rawMessage, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		trimmedMessage := strings.TrimSpace(rawMessage)
		splitMessage := strings.SplitN(trimmedMessage, " ", 2)
		msg := message.MakeMutableMessage(splitMessage[0])

		if len(splitMessage) > 1 {
			msg.Append(splitMessage[1])
		}

		msg.Finalize()

		eng.Do(&associatedObj, msg)
	}
}
