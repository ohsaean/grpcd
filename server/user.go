package main

import (
	"github.com/ohsaean/grpcd/lib"
	pb "github.com/ohsaean/grpcd/proto"
	"time"
)

type User struct {
	userID int64
	room   *Room
	recv   chan *pb.Message
	exit   chan bool // signal
}

func NewUser(uid int64, room *Room) *User {
	return &User{
		userID: uid,
		recv:   make(chan *pb.Message),
		exit:   make(chan bool, 1),
		room:   room,
	}
}

func NewRootMessage(userID int64) *pb.Message {
	return &pb.Message{
		UserID:    userID,
		Timestamp: time.Now().Unix(),
	}
}

func (u *User) Leave() {

	lib.Log("Leave user id : ", lib.Itoa64(u.userID))

	if u.room == nil {
		lib.Log("Error, room is nil")
		return
	}

	lib.Log("Leave room id : ", lib.Itoa64(u.room.roomID))

	// broadcast message
	notifyMsg := NewRootMessage(u.userID)
	notifyMsg.Payload = &pb.Message_NotifyQuit{
		NotifyQuit: &pb.NotifyQuitMsg{
			RoomID: u.room.roomID,
		},
	}

	u.SendToAll(notifyMsg)

	u.room.Leave(u.userID)

	sessions.Remove(u.userID)

	lib.Log("NotifyQuit message send")

	lib.Log("Leave func end")
}

func (u *User) Push(m *pb.Message) {
	u.recv <- m // send message to user
}

func (u *User) SendToAll(m *pb.Message) {
	if u.room.IsEmptyRoom() == false {
		u.room.messages <- m
	}
}
