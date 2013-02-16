package engine

import (
	"github.com/jworrell/GoMU/database"
	"github.com/jworrell/GoMU/message"
	"github.com/jworrell/GoMU/object"
)

type Engine struct {
	db *database.Database
}

func Init(path string) (*Engine, error) {
	db, err := database.LoadDB(path)
	if err != nil {
		return nil, err
	}
	return &Engine{db}, nil
}

func (eng *Engine) Do(obj **object.Object, msg *message.Message) {
	cmd := Commands[msg.Command]

	if cmd.function == nil {
		obj.Hear(&message.Message{"error", msg.Command + " is not a valid command!"})
	} else if obj.GetID() != object.DUMMY_ID || cmd.UseUnathenticated() {
		cmd.function(eng, obj, msg)
	} else {
		obj.Hear(&message.Message{"error", "You must be logged on to use " + msg.Command})
	}

}
