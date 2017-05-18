package main

import (
	"github.com/ohsaean/grpcd/lib"
	pb "github.com/ohsaean/grpcd/proto"
)

func loginHandler(in *pb.Message, user *User) {

	req := in.GetReqLogin()
	if req == nil {
		lib.Log("fail, GetReqLogin()")
	} else {
		lib.Log("GetReqLogin() : ", in)
	}
	user.userID = in.UserID

	// TODO validation logic here

	res := NewRootMessage(user.userID)
	res.Payload = &pb.Message_ResLogin{
		ResLogin: &pb.ResLogin{},
	}

	user.Push(res)
}

func createHandler(in *pb.Message, user *User) {

	req := in.GetReqCreate()
	if req == nil {
		lib.Log("fail, GetReqCreate()")
	}

	lib.Log("GetReqCreate() in : ", in)
	lib.Log("GetReqCreate() user : ", user)

	if user.userID != in.UserID {
		lib.Log("Fail room create, user id missmatch")
		return
	}

	// room create
	roomID := GetAutoIncRoomID()
	r := NewRoom(roomID)
	r.users.Set(user.userID, user) // insert user
	user.room = r                  // set room
	lib.Log("user ", user)
	rooms.Set(roomID, r) // set room into global shared map
	lib.Log("Get rand room id : ", lib.Itoa64(roomID))

	res := NewRootMessage(user.userID)
	res.Payload = &pb.Message_ResCreate{
		ResCreate: &pb.ResCreate{
			RoomID: roomID,
		},
	}

	lib.Log("Room create, room id : ", lib.Itoa64(roomID))
	user.Push(res)
	return
}

func joinHandler(in *pb.Message, user *User) {

	// request body unmarshaling
	req := in.GetReqJoin()
	if req == nil {
		lib.Log("fail, GetReqJoin()")
	}

	lib.Log("GetReqJoin() : ", in)

	roomID := req.RoomID

	value, ok := rooms.Get(roomID)

	if !ok {
		lib.Log("Fail room join, room does not exist, room id : ", lib.Itoa64(roomID))
		return
	}

	r := value.(*Room)
	r.users.Set(user.userID, user)
	user.room = r

	// broadcast message
	notifyMsg := NewRootMessage(user.userID)
	notifyMsg.Payload = &pb.Message_ReqJoin{
		ReqJoin: &pb.ReqJoin{
			RoomID: roomID,
		},
	}

	user.SendToAll(notifyMsg)

	// response body marshaling
	res := NewRootMessage(user.userID)
	res.Payload = &pb.Message_ResJoin{
		ResJoin: &pb.ResJoin{
			RoomID: roomID,
		},
	}

	user.Push(res)
}

func action1Handler(in *pb.Message, user *User) {

	// request body unmarshaling
	req := in.GetReqAction1()
	if req == nil {
		lib.Log("fail, GetReqAction1()")
	}

	// TODO create business logic for Action1 Type

	lib.Log("Action1 userID : ", in)

	// broadcast message
	notifyMsg := NewRootMessage(user.userID)
	notifyMsg.Payload = &pb.Message_NotifyAction1{
		NotifyAction1: &pb.NotifyAction1Msg{},
	}

	user.SendToAll(notifyMsg)

	// response body marshaling
	res := NewRootMessage(user.userID)
	res.Payload = &pb.Message_ResAction1{
		ResAction1: &pb.ResAction1{
			Result: 1,
		},
	}
	user.Push(res)
}

func quitHandler(in *pb.Message, user *User) {

	// request body unmarshaling
	req := in.GetReqQuit()
	if req == nil {
		lib.Log("fail, GetReqQuit()")
	}

	// response body marshaling
	res := NewRootMessage(user.userID)
	res.Payload = &pb.Message_ResQuit{
		ResQuit: &pb.ResQuit{},
	}
	user.Push(res)

	// same act user.Leave()
	user.exit <- true
}

func roomListHandler(in *pb.Message, user *User) {
	// request body unmarshaling
	req := in.GetReqRoomList()
	if req == nil {
		lib.Log("fail, GetReqQuit()")
	}
	lib.Log("GetReqRoomList() : ", in)

	// response body marshaling
	res := NewRootMessage(user.userID)
	res.Payload = &pb.Message_ResRoomList{
		ResRoomList: &pb.ResRoomList{
			RoomIDs: rooms.GetKeys(),
		},
	}

	user.Push(res)
}
