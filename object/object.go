package object

import (
	"github.com/jworrell/GoMU/message"
	"log"
	"sort"
	"strings"
	"sync"
)

type ObjectType uint8
type ObjectID uint32

const (
	_ ObjectType = iota
	ROOM
	PLAYER
	THING
	EXIT
)

const (
	DUMMY_ID     ObjectID = 0
	NIL_LOCATION ObjectID = 0
)

type SerializableObject struct {
	ID         ObjectID
	Kind       ObjectType
	Owner      ObjectID
	Home       ObjectID
	Links      ObjectID
	Attributes map[string]string
}

type Object struct {
	sync.RWMutex

	id       ObjectID
	kind     ObjectType
	owner    *Object
	home     *Object
	location *Object
	links    *Object

	contents map[*Object]bool

	attributes map[string]string

	writer chan *message.Message
	saver  chan *SerializableObject
}

func NewObject(saver chan *SerializableObject, id ObjectID) *Object {
	o := Object{}
	o.id = id

	o.contents = make(map[*Object]bool)
	o.attributes = make(map[string]string)

	o.saver = saver
	o.save()

	return &o
}

// Not thread safe, only call from thread safe functions
func (obj *Object) save() {
	if obj.saver != nil {
		so := obj.serialize()
		obj.saver <- so
	}
}

func (obj *Object) saveAndUnlock() {
	obj.save()
	obj.Unlock()
}

// Thread safe wrapper for serialize
func (obj *Object) Serialize() *SerializableObject {
	obj.RLock()
	defer obj.RUnlock()

	return obj.serialize()
}

// Not thread safe, only call from thread safe functions
func (obj *Object) serialize() *SerializableObject {
	homeId := ObjectID(0)
	linksId := ObjectID(0)
	ownerId := ObjectID(0)

	if obj.home != nil {
		homeId = obj.home.id
	}

	if obj.links != nil {
		linksId = obj.links.id
	}

	if obj.owner != nil {
		ownerId = obj.owner.id
	}

	so := SerializableObject{
		obj.id,
		obj.kind,
		ownerId,
		homeId,
		linksId,
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

func (obj *Object) GetContents() ObjectSlice {
	obj.RLock()
	defer obj.RUnlock()

	contents := make([]*Object, len(obj.contents))
	idx := 0

	for o := range obj.contents {
		contents[idx] = o
		idx++
	}

	return contents
}

func (obj *Object) GetOwner() *Object {
	obj.RLock()
	defer obj.RUnlock()

	return obj.owner
}

func (obj *Object) SetOwner(owner *Object) {
	obj.Lock()
	defer obj.saveAndUnlock()

	obj.owner = owner
}

func (obj *Object) GetLink() *Object {
	obj.RLock()
	defer obj.RUnlock()

	return obj.links
}

func (obj *Object) SetLink(link *Object) {
	obj.Lock()
	defer obj.saveAndUnlock()

	obj.links = link
}

func (obj *Object) GetHome() *Object {
	obj.RLock()
	defer obj.RUnlock()

	return obj.home
}

func (obj *Object) SetHome(home *Object) {
	obj.Lock()
	defer obj.saveAndUnlock()

	obj.home = home
}

func (obj *Object) GetType() ObjectType {
	obj.RLock()
	defer obj.RUnlock()

	return obj.kind
}

func (obj *Object) SetType(kind ObjectType) {
	obj.Lock()
	defer obj.saveAndUnlock()

	obj.kind = kind
}

func (obj *Object) GetAttr(attr string) string {
	obj.RLock()
	defer obj.RUnlock()

	return obj.attributes[attr]
}

func (obj *Object) SetAttr(attr, value string) {
	obj.Lock()
	defer obj.saveAndUnlock()

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

	obj.writer = writer
}

func (obj *Object) Hear(msg *message.Message) {
	obj.RLock()
	defer obj.RUnlock()

	if obj.writer != nil {
		select {
		case obj.writer <- msg:
		default:
			log.Println("Dropped a message for " + obj.attributes["name"])
		}
	}
}

func (obj *Object) FindInside(key, pattern string) *Object {
	if strings.Contains(strings.ToLower(obj.GetAttr(key)), strings.ToLower(pattern)) {
		return obj
	}

	contents := obj.GetContents()

	for _, o := range contents {
		if strings.Contains(strings.ToLower(o.GetAttr(key)), strings.ToLower(pattern)) {
			return o
		}
	}

	return nil
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

type ObjectSliceByType []*Object

func (os ObjectSliceByType) Len() int {
	return len(os)
}

func (os ObjectSliceByType) Less(i, j int) bool {
	return os[i].GetType() < os[j].GetType()
}

func (os ObjectSliceByType) Swap(i, j int) {
	tmp := os[i]
	os[i] = os[j]
	os[j] = tmp
}
