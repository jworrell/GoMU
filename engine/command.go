package engine

import (
	"github.com/jworrell/GoMU/message"
	"github.com/jworrell/GoMU/object"
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
	"pose":  Command{(*Engine).pose, false},
	"look":  Command{(*Engine).look, false},
	"l":     Command{(*Engine).look, false},
	"move":  Command{(*Engine).move, false},
	"m":     Command{(*Engine).move, false},
}

func (eng *Engine) login(obj **object.Object, msg *message.Message) {
	player := eng.db.GetPlayer(msg.Data)

	if player == nil {
		obj.Hear(message.MakeMessage("error", "No player named "+msg.Data+" exists!"))
	} else {
		writer := obj.GetWriter()
		player.SetWriter(writer)
		player.Hear(message.MakeMessage("result", "Successfully logged in as "+player.GetAttr("name")))
		(*obj) = player
		player.GetLocation().Emit(message.MakeMessage("emit", player.GetAttr("name")+" has connected!"))
		eng.look(obj, &message.Message{})
	}
}

func (eng *Engine) say(obj **object.Object, msg *message.Message) {
	obj.GetLocation().Emit(message.MakeMessage("emit", obj.GetAttr("name")+" says, \""+msg.Data+"\""))
}

func (eng *Engine) pose(obj **object.Object, msg *message.Message) {
	obj.GetLocation().Emit(message.MakeMessage("emit", obj.GetAttr("name")+" "+msg.Data))
}

func (eng *Engine) look(obj **object.Object, msg *message.Message) {
	location := obj.GetLocation()
	result := message.MakeMutableMessage("result")

	if msg.Data == "" {
		contents := location.GetContents()
		sort.Sort(object.ObjectSliceByType(contents))

		result.Append(location.GetAttr("name") + "\n" + location.GetAttr("description") + "\n\n")
		result.Append("Contents:\n")

		for _, thing := range contents {
			if thing.GetType() != object.PLAYER || thing.GetWriter() != nil {
				result.Append(thing.GetAttr("name") + "\n")
			}
		}
	} else {
		subj := location.FindInside("name", msg.Data)
		if subj == nil {
			result.Append(msg.Data + " not found!")
		} else {
			result.Append(subj.GetAttr("name") + "\n" + subj.GetAttr("description") + "\n\n")
		}
	}

	obj.Hear(result.Finalize())
}

func (eng *Engine) move(obj **object.Object, msg *message.Message) {
	location := obj.GetLocation()
	exit := location.FindInside("name", msg.Data)

	if exit != nil && exit.GetType() == object.EXIT {
		dest := exit.GetLink()
		location.Emit(message.MakeMessage("emit", obj.GetAttr("name")+" left through "+exit.GetAttr("name")))
		obj.Move(dest)
		dest.Emit(message.MakeMessage("emit", obj.GetAttr("name")+" arrived from "+location.GetAttr("name")))
		eng.look(obj, &message.Message{})
	} else {
		obj.Hear(message.MakeMessage("error", msg.Data+" not found or not an exit!"))
	}
}
