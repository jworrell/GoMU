package database

import (
	"code.google.com/p/gosqlite/sqlite"
	"encoding/json"
	"github.com/jworrell/GoMU/object"
	"io"
	"log"
	"os"
	"sync"
)

const (
	WRITE_QUEUE_LENGTH = 1024
)

type Database struct {
	sync.RWMutex
	objects map[object.ObjectID]*object.Object
	players map[string]*object.Object
	saver   chan *object.SerializableObject
	nextId  object.ObjectID
}

func InitDB(path string) (*Database, error) {
	db := Database{
		sync.RWMutex{},
		make(map[object.ObjectID]*object.Object),
		make(map[string]*object.Object),
		make(chan *object.SerializableObject, WRITE_QUEUE_LENGTH),
		0,
	}

	sqliteDb, err := sqlite.Open(path)
	if err != nil {
		return nil, err
	}

	go func() {
		defer sqliteDb.Close()

		insertStmnt, err := sqliteDb.Prepare("INSERT OR REPLACE INTO objects (id, data) VALUES (?, ?)")
		if err != nil {
			panic("Failed to create insert statement. This shouldn't happen!")
		}

		for {
			so := <-db.saver
			sob, err := json.Marshal(so)
			if err != nil {
				log.Println(sob)
				continue
			}

			err = insertStmnt.Exec(so.ID, sob)
			if err != nil {
				log.Println(sob)
				continue
			}

			err = insertStmnt.Next()
			if err != nil {
				log.Println(sob)
				continue
			}
		}
	}()

	selectStmnt, err := sqliteDb.Prepare("SELECT data FROM objects")
	if err != nil {
		return nil, err
	}

	err = selectStmnt.Exec()
	if err != nil {
		return nil, err
	}

	// TODO: Right now, it reads objects from the DB and then reinserts them. Figure out a way to fix this that doesn't suck.
	for selectStmnt.Next() {
		jsonObj := make([]byte, 0)
		selectStmnt.Scan(&jsonObj)
		so := &object.SerializableObject{}
		json.Unmarshal(jsonObj, so)
		db.AddSerializableObject(so)
		if so.ID >= db.nextId {
			db.nextId++
		}
	}

	return &db, nil
}

func (db *Database) LoadJSON(path string) error {
	so := &object.SerializableObject{}

	inFile, err := os.Open(path)

	if err != nil {
		return err
	}

	decoder := json.NewDecoder(inFile)

	for {
		err := decoder.Decode(so)

		if err != nil {
			if err == io.EOF {
				return err
			}
		}

		db.AddSerializableObject(so)
	}

	return nil
}

func (db *Database) AddSerializableObject(so *object.SerializableObject) {
	workingObj := db.getOrCreateObj(so.ID)
	workingObj.SetType(so.Kind)
	workingObj.SetOwner(db.getOrCreateObj(so.Owner))

	if so.Kind == object.PLAYER {
		db.players[so.Attributes["name"]] = workingObj
	}

	if so.Home != object.NIL_LOCATION {
		workingObj.SetHome(db.getOrCreateObj(so.Home))
		workingObj.Move(workingObj.GetHome())
	}

	if so.Home != object.NIL_LOCATION {
		workingObj.SetLink(db.getOrCreateObj(so.Links))
	}

	for k, v := range so.Attributes {
		workingObj.SetAttr(k, v)
	}
}

func (db *Database) CreateObject() *object.Object {
	db.Lock()
	defer db.Unlock()

	id := db.getNextId()
	obj := object.NewObject(db.saver, id)

	return obj
}

func (db *Database) GetPlayer(name string) *object.Object {
	db.RLock()
	defer db.RUnlock()

	return db.players[name]
}

func (db *Database) GetObject(id object.ObjectID) *object.Object {
	db.RLock()
	defer db.RUnlock()

	return db.objects[id]
}

func (db *Database) getOrCreateObj(id object.ObjectID) *object.Object {
	db.Lock()
	defer db.Unlock()

	obj := db.objects[id]

	if obj == nil {
		obj = object.NewObject(db.saver, id)
		db.objects[id] = obj
	}

	return obj
}

// Not thread safe
func (db *Database) getNextId() object.ObjectID {
	id := db.nextId
	db.nextId++
	return id
}
