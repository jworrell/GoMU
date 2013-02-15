package object

import (
	"GoMU/message"
	"sort"
	"sync"
)

type ObjectType uint8
type ObjectID uint32

const (
	_ ObjectType = iota
	ROOM
	PLAYER
	EXIT
	THING
)

const (
	DUMMY_ID     ObjectID = 0
	NIL_LOCATION ObjectID = 0
)

type SerializeableObject struct {
	ID         ObjectID
	Kind       ObjectType
	Owner      ObjectID
	Home       ObjectID
	Attributes map[string]string
}

type Object struct {
	sync.RWMutex

	id       ObjectID
	kind     ObjectType
	dirty    bool
	owner    *Object
	home     *Object
	location *Object

	contents map[*Object]bool

	attributes map[string]string

	writer chan *message.Message
}

func NewObject(id ObjectID) *Object {
	o := Object{}
	o.id = id

	o.contents = make(map[*Object]bool)
	o.attributes = make(map[string]string)

	return &o
}

func (obj *Object) Serialize() *SerializeableObject {
	obj.RLock()
	defer obj.RUnlock()

	homeId := ObjectID(0)

	if obj.home != nil {
		homeId = obj.home.id
	}

	so := SerializeableObject{
		obj.id,
		obj.kind,
		obj.owner.id,
		homeId,
		make(map[string]string),
	}

	for k, v := range obj.attributes {
		so.Attributes[k] = v
	}

	return &so
}

func (obj *Object) GetID() ObjectID {
	return obj.id
}

func (obj *Object) GetOwner() *Object {
	obj.RLock()
	defer obj.RUnlock()

	return obj.owner
}

func (obj *Object) SetOwner(owner *Object) {
	obj.Lock()
	defer obj.Unlock()

	obj.dirty = true
	obj.owner = owner
}

func (obj *Object) GetHome() *Object {
	obj.RLock()
	defer obj.RUnlock()

	return obj.home
}

func (obj *Object) SetHome(home *Object) {
	obj.Lock()
	defer obj.Unlock()

	obj.dirty = true
	obj.home = home
}

func (obj *Object) GetType() ObjectType {
	obj.RLock()
	defer obj.RUnlock()

	return obj.kind
}

func (obj *Object) SetType(kind ObjectType) {
	obj.Lock()
	defer obj.Unlock()

	obj.dirty = true
	obj.kind = kind
}

func (obj *Object) GetAttr(attr string) string {
	obj.RLock()
	defer obj.RUnlock()

	return obj.attributes[attr]
}

func (obj *Object) SetAttr(attr, value string) {
	obj.Lock()
	defer obj.Unlock()

	obj.dirty = true
	obj.attributes[attr] = value
}

func (obj *Object) GetLocation() *Object {
	obj.RLock()
	defer obj.RUnlock()

	return obj.location
}

func (obj *Object) Move(dest *Object) bool {
	var lockList ObjectSlice

	loc := obj.GetLocation()

	if loc == nil {
		lockList = ObjectSlice{obj, dest}
	} else {
		lockList = ObjectSlice{obj, loc, dest}
	}

	sort.Sort(lockList)

	for _, o := range lockList {
		o.Lock()
		defer o.Unlock()
	}

	if obj.location != loc {
		return false
	}

	obj.location = dest
	dest.contents[obj] = true

	if loc != nil {
		delete(loc.contents, obj)
	}

	return true
}

func (obj *Object) GetWriter() chan *message.Message {
	obj.RLock()
	defer obj.RUnlock()

	return obj.writer
}

func (obj *Object) SetWriter(writer chan *message.Message) {
	obj.Lock()
	defer obj.Unlock()

	obj.dirty = true
	obj.writer = writer
}

func (obj *Object) Hear(msg *message.Message) {
	obj.RLock()

	if obj.writer != nil {
		select {
		case obj.writer <- msg:
			obj.RUnlock()

		default:
			obj.RUnlock()

			obj.Lock()
			obj.writer = nil
			obj.Unlock()
		}
	}
}

func (obj *Object) Emit(msg *message.Message) {
	obj.RLock()
	defer obj.RUnlock()

	for o := range obj.contents {
		o.Hear(msg)
	}
}

type ObjectSlice []*Object

func (os ObjectSlice) Len() int {
	return len(os)
}

func (os ObjectSlice) Less(i, j int) bool {
	return os[i].id < os[j].id
}

func (os ObjectSlice) Swap(i, j int) {
	tmp := os[i]
	os[i] = os[j]
	os[j] = tmp
}
