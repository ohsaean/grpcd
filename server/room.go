package main

import (
	"github.com/ohsaean/grpcd/lib"
	pb "github.com/ohsaean/grpcd/proto"
	"sync"
)

type Room struct {
	users    lib.SharedMap
	messages chan *pb.Message // broadcast message channel
	roomID   int64
}

var (
	lastUseRoomID int64
	sc            sync.RWMutex
)

func GetAutoIncRoomID() int64 {
	sc.Lock()
	defer sc.Unlock()
	lastUseRoomID += 1
	return lastUseRoomID
}

func GetRoom(roomID int64) (r *Room) {
	value, ok := rooms.Get(roomID)
	if !ok {
		lib.Log("err: not exist room : ", roomID)
	}
	r = value.(*Room)
	return
}

// Room construct
func NewRoom(roomID int64) (r *Room) {
	r = new(Room)
	r.messages = make(chan *pb.Message)
	r.users = lib.NewSMap(lib.RWMutex) // global shared map
	r.roomID = roomID
	go r.RoomMessageLoop() // for broadcast message
	return
}

func (r *Room) RoomMessageLoop() {
	// when messages channel is closed then "for-loop" will be break
	for m := range r.messages {
		for userID := range r.users.Map() {
			if userID == m.UserID {
				continue
			}
			value, ok := r.users.Get(userID)
			if !ok {
				continue
			}
			user := value.(*User)
			user.Push(m)
		}
	}
}

func (r *Room) getRoomUsers() (userIDs []int64) {
	userIDs = r.users.GetKeys()
	return userIDs
}

func (r *Room) Leave(userID int64) {
	r.users.Remove(userID)

	if r.IsEmptyRoom() == false {
		return
	}

	close(r.messages) // close broadcast channel
	rooms.Remove(r.roomID)
}

func (r *Room) IsEmptyRoom() bool {
	if r.users.Count() == 0 {
		return true
	}
	return false
}
