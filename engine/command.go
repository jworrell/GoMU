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
	"login":    Command{(*Engine).login, true},
	"register": Command{(*Engine).register, true},
	"say":      Command{(*Engine).say, false},
	"wall":     Command{(*Engine).wall, false},
	"pose":     Command{(*Engine).pose, false},
	"look":     Command{(*Engine).look, false},
	"l":        Command{(*Engine).look, false},
	"move":     Command{(*Engine).move, false},
	"m":        Command{(*Engine).move, false},
	"shutdown": Command{(*Engine).shutdown, false},
}

func (eng *Engine) login(obj **object.Object, msg *message.Message) {
	player := eng.db.GetPlayer(msg.Data)

	if player == nil {
		obj.Hear(message.MakeMessage("error", "No player named %s exists!", msg.Data))
	} else {
		writer := obj.GetWriter()
		player.SetWriter(writer)
		player.Hear(message.MakeMessage("result", "Successfully logged in as %s", player.GetAttr("name")))
		(*obj) = player
		player.GetLocation().Emit(message.MakeMessage("emit", "%s has connected!", player.GetAttr("name")))
		eng.look(obj, &message.Message{})
	}
}

func (eng *Engine) register(obj **object.Object, msg *message.Message) {
	player := eng.db.GetPlayer(msg.Data)

	if player != nil {
		obj.Hear(message.MakeMessage("error", "Player named %s already exists!", msg.Data))
	} else {
		writer := obj.GetWriter()
		newPlayer := eng.db.CreateObject()
		newPlayer.SetWriter(writer)
		newPlayer.SetAttr("name", msg.Data)
		newPlayer.SetType(object.PLAYER)
		newPlayer.SetOwner(newPlayer)

		location := eng.db.GetObject(object.DEFAULT_LOCATION)
		newPlayer.SetHome(location)
		newPlayer.Move(location)

		newPlayer.Hear(message.MakeMessage("result", "Successfully registered as %s", msg.Data))

		(*obj) = newPlayer
		newPlayer.GetLocation().Emit(message.MakeMessage("emit", "%s has connected!", newPlayer.GetAttr("name")))
		eng.look(obj, &message.Message{})
	}
}

func (eng *Engine) say(obj **object.Object, msg *message.Message) {
	obj.GetLocation().Emit(message.MakeMessage("emit", "%s says, \"%s\"", obj.GetAttr("name"), msg.Data))
}

func (eng *Engine) pose(obj **object.Object, msg *message.Message) {
	obj.GetLocation().Emit(message.MakeMessage("emit", "%s %s", obj.GetAttr("name"), msg.Data))
}

func (eng *Engine) look(obj **object.Object, msg *message.Message) {
	location := obj.GetLocation()
	result := message.MakeMutableMessage("result")

	if msg.Data == "" {
		contents := location.GetContents()
		sort.Sort(object.ObjectSliceByType(contents))

		result.Append("%s (#%d)\n%s\n\n", location.GetAttr("name"), location.GetID(), location.GetAttr("description"))
		result.Append("Contents:\n")

		for _, thing := range contents {
			if thing.GetType() != object.PLAYER || thing.GetWriter() != nil {
				result.Append("%s (#%d)\n", thing.GetAttr("name"), thing.GetID())
			}
		}
	} else {
		subj := location.FindInside("name", msg.Data)
		if subj == nil {
			result.Append("%s not found!", msg.Data)
		} else {
			result.Append("%s (#%d)\n%s\n", subj.GetAttr("name"), subj.GetID(), subj.GetAttr("description"))
		}
	}

	obj.Hear(result.Finalize())
}

func (eng *Engine) move(obj **object.Object, msg *message.Message) {
	location := obj.GetLocation()
	exit := location.FindInside("name", msg.Data)

	if exit != nil && exit.GetType() == object.EXIT {
		dest := exit.GetLink()
		location.Emit(message.MakeMessage("emit", "%s left through %s", obj.GetAttr("name"), exit.GetAttr("name")))
		obj.Move(dest)
		dest.Emit(message.MakeMessage("emit", "%s arrived from %s", obj.GetAttr("name"), location.GetAttr("name")))
		eng.look(obj, &message.Message{})
	} else {
		obj.Hear(message.MakeMessage("error", "%s not found or not an exit!", msg.Data))
	}
}

func (eng *Engine) wall(obj **object.Object, msg *message.Message) {
	for _, player := range eng.db.GetPlayers() {
		player.Hear(msg)
	}
}

func (eng *Engine) shutdown(obj **object.Object, msg *message.Message) {
	eng.wall(obj, message.MakeMessage("emit", "Server is shutting down now!"))
	eng.Shutdown()
}
