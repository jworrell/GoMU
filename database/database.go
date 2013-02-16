package database

import (
	"github.com/jworrell/GoMU/object"
	"encoding/json"
	"os"
	"sync"
	"io"
)

type Database struct {
	sync.RWMutex
	objects map[object.ObjectID]*object.Object
	players map[string]*object.Object
}

func LoadDB(path string) (*Database,error) {
	db := Database{
		sync.RWMutex{},
		make(map[object.ObjectID]*object.Object),
		make(map[string]*object.Object),
	}

	so := &object.SerializeableObject{}

	inFile, err := os.Open(path)

	if err != nil {
		return nil,err
	}

	decoder := json.NewDecoder(inFile)

	for {
		err := decoder.Decode(so)
		
		if err != nil {
			if err == io.EOF {break}
			return nil,err
		}

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

	return &db,nil
}

func (db *Database) getOrCreateObj(id object.ObjectID) *object.Object {
	db.Lock()
	defer db.Unlock()

	obj := db.objects[id]

	if obj == nil {
		obj = object.NewObject(id)
		db.objects[id] = obj
	}

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
