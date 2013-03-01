package engine

import (
	"github.com/jworrell/GoMU/database"
	"github.com/jworrell/GoMU/message"
	"github.com/jworrell/GoMU/object"
	"os"
	"sync"
)

var engineCount int
var engineMutex sync.Mutex

type Engine struct {
	db *database.Database
}

func Init(path string) (*Engine, error) {
	var err error
	var db *database.Database

	engineMutex.Lock()
	defer engineMutex.Unlock()

	db, err = database.InitDB(path)
	if err != nil {
		return nil, err
	}

	engineCount++

	return &Engine{db}, nil
}

func (eng *Engine) Do(obj **object.Object, msg *message.Message) {
	cmd := Commands[msg.Command]

	if cmd.function == nil {
		obj.Hear(message.MakeMessage("error", msg.Command+" is not a valid command!"))
	} else if obj.GetID() != object.DUMMY_ID || cmd.UseUnathenticated() {
		cmd.function(eng, obj, msg)
	} else {
		obj.Hear(message.MakeMessage("error", "You must be logged on to use "+msg.Command))
	}
}

func (eng *Engine) Shutdown() {
	engineMutex.Lock()
	defer engineMutex.Unlock()

	eng.db.Shutdown()

	// When the last engine shuts down, exit the server
	engineCount--
	if engineCount <= 0 {
		os.Exit(0)
	}
}
