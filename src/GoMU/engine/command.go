package engine

import (
	"GoMU/message"
	"GoMU/object"
	"sort"
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
	"look":  Command{(*Engine).look, false},
	"l":     Command{(*Engine).look, false},
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

func (eng *Engine) look(obj **object.Object, msg *message.Message) {
	location := obj.GetLocation()
	contents := location.GetContents()
	sort.Sort(object.ObjectSliceByType(contents))

	buffer := location.GetAttr("name") + ": " + location.GetAttr("description") + "\n\n"
	buffer += "Contents:\n"

	for _, thing := range contents {
		if thing.GetType() != object.PLAYER || thing.GetWriter() != nil {
			buffer += thing.GetAttr("name") + "\n"
		}
	}

	obj.Hear(&message.Message{"result", buffer})
}
