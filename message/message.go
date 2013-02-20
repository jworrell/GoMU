package message

import (
	"fmt"
)

type Message struct {
	Command   string
	Data      string
	buffer    []byte
	finalized bool
}

func MakeMessage(cmd, format string, data ...interface{}) *Message {
	formattedString := fmt.Sprintf(format, data...)
	return &Message{cmd, formattedString, make([]byte, 0), true}
}

func MakeMutableMessage(cmd string) *Message {
	return &Message{cmd, "", make([]byte, 0), false}
}

func (msg *Message) Append(format string, data ...interface{}) {
	if msg.finalized {
		panic("Attempt to apend to finalized message")
	}

	formattedString := fmt.Sprintf(format, data...)
	msg.buffer = append(msg.buffer, []byte(formattedString)...)
}

func (msg *Message) Finalize() *Message {
	if msg.buffer != nil {
		msg.Data = string(msg.buffer)
	}

	msg.finalized = true
	return msg
}

func (msg *Message) IsFinalized() bool {
	return msg.finalized
}
