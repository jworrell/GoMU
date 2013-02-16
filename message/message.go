package message

type Message struct {
	Command   string
	Data      string
	buffer    []byte
	finalized bool
}

func MakeMessage(cmd, data string) *Message {
	return &Message{cmd, data, make([]byte, 0), true}
}

func MakeMutableMessage(cmd string) *Message {
	return &Message{cmd, "", make([]byte, 0), false}
}

func (msg *Message) Append(data string) {
	if msg.finalized {
		panic("Attempt to apend to finalized message")
	}

	msg.buffer = append(msg.buffer, []byte(data)...)
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
