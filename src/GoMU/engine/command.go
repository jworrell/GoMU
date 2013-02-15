package engine

import (
	"GoMU/message"
	"GoMU/object"
)

type Command struct {
	function       func(*Engine, **object.Object, *message.Message)
	unathenticated bool
}

func (com *Command) UseUnathenticated() bool {
	return com.unathenticated
}

var Commands = map[string]Command{
	"login": Command{(*Engine).login, true},
	"say":   Command{(*Engine).say, false},
}

func (eng *Engine) login(obj **object.Object, msg *message.Message) {
	player := eng.db.GetPlayer(msg.Data)

	if player == nil {
		obj.Hear(&message.Message{"error", "No player named " + msg.Data + " exists!"})
	} else {
		player.SetWriter(obj.GetWriter())
		player.Hear(&message.Message{"result", "Successsss!!"})
		(*obj) = player
	}
}

func (eng *Engine) say(obj **object.Object, msg *message.Message) {
	obj.GetLocation().Emit(msg)
}
