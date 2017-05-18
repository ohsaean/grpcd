package main

import (
	"github.com/lxn/walk"
	walk_dcl "github.com/lxn/walk/declarative"
	"github.com/ohsaean/grpcd/lib"
	pb "github.com/ohsaean/grpcd/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"log"
	"os"
	"strconv"
	"time"
)

var inputString string

func main() {

	var stream pb.Gateway_RouteMessageClient

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial("localhost:50051", opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewGatewayClient(conn)

	stream, err = client.RouteMessage(context.Background())
	if err != nil {
		log.Fatalf("%v.RouteMessage(_) = _, %v", client, err)
	}

	mw := createMainForm(stream)

	go func() {
		for {
			log.Println("wait for read")
			in, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				return
			}
			messageHandler(in)
		}
	}()

	mw.Run()
}

func createMainForm(stream pb.Gateway_RouteMessageClient) (mw *walk.MainWindow) {
	var userID int64
	err := walk_dcl.MainWindow{
		AssignTo: &mw,
		Title:    "Grpcd Client",
		MinSize: walk_dcl.Size{
			Width:  320,
			Height: 240,
		},
		Size: walk_dcl.Size{
			Width:  600,
			Height: 400,
		},
		Layout: walk_dcl.VBox{},
		Children: []walk_dcl.Widget{
			walk_dcl.HSplitter{
				Children: []walk_dcl.Widget{
					walk_dcl.PushButton{
						Text: "login",
						OnClicked: func() {
							if cmd, err := RunUserIdDialog(mw); err != nil {
								log.Print(err)
							} else if cmd == walk.DlgCmdOK {
								log.Println("dlg msg : " + inputString)
								num, err := strconv.Atoi(inputString)
								lib.CheckError(err)
								userID = int64(num)
								ReqLogin(stream, userID)
							}
						},
					},
					walk_dcl.PushButton{
						Text: "room create",
						OnClicked: func() {
							log.Println("req create user id : ", userID)
							ReqCreate(stream, userID)
						},
					},
					walk_dcl.PushButton{
						Text: "room list",
						OnClicked: func() {
							log.Println("room list user id : ", userID)
							ReqRoomList(stream, userID)
						},
					},
					walk_dcl.PushButton{
						Text: "join",
						OnClicked: func() {
							if cmd, err := RunRoomJoinDialog(mw); err != nil {
								log.Print(err)
							} else if cmd == walk.DlgCmdOK {
								log.Println("dlg msg : " + inputString)
								num, err := strconv.Atoi(inputString)
								lib.CheckError(err)
								roomID := int64(num)
								ReqJoin(stream, userID, roomID)
							}
						},
					},
					walk_dcl.PushButton{
						Text: "action1",
						OnClicked: func() {
							ReqAction1(stream, userID)
						},
					},
					walk_dcl.PushButton{
						Text: "quit",
						OnClicked: func() {
							log.Println("quit user id : ", userID)
							ReqQuit(stream, userID)
							os.Exit(3)
						},
					},
				},
			},
		},
	}.Create()

	if err != nil {
		log.Fatal(err)
	}

	lv, err := NewLogView(mw)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(lv)

	return
}

// MessageHandler 여기서 각 proto message 에 대한 적절한 프로시저를 할당함
func messageHandler(msg *pb.Message) {
	// type switch 말고는 방법이 없나??
	switch msg.Payload.(type) {

	case *pb.Message_ResLogin:
		ResLogin(msg)

	case *pb.Message_ResCreate:
		ResCreate(msg)

	case *pb.Message_ResJoin:
		ResJoin(msg)

	case *pb.Message_ResAction1:
		ResAction1(msg)

	case *pb.Message_ResRoomList:
		ResRoomList(msg)

	case *pb.Message_ResQuit:
		ResQuit(msg)

	case *pb.Message_NotifyAction1:
		NotifyAction1Handler(msg)

	case *pb.Message_NotifyJoin:
		NotifyJoinHandler(msg)

	case *pb.Message_NotifyQuit:
		NotifyQuitHandler(msg)

	default:
		log.Println("Error, not defined handler")
	}
}

func NewRootMessage(userID int64) *pb.Message {
	return &pb.Message{
		UserID:    userID,
		Timestamp: time.Now().Unix(),
	}
}

func ReqLogin(stream pb.Gateway_RouteMessageClient, userUID int64) {

	req := NewRootMessage(userUID)
	req.Payload = &pb.Message_ReqLogin{
		ReqLogin: &pb.ReqLogin{},
	}

	stream.Send(req)
	log.Println("ReqLogin client send :", req)
}

func ResLogin(data *pb.Message) bool {
	res := data.GetResLogin()
	if res == nil {
		log.Println("fail, GetResLogin()")
		return false
	} else {
		log.Println("GetResLogin() : ", res)
	}

	log.Println("ResLogin server return : user id : " + lib.Itoa64(data.UserID))
	log.Println("ResLogin server return : result code : " + lib.Itoa32(res.Result))
	return true
}

func ReqRoomList(stream pb.Gateway_RouteMessageClient, userUID int64) {
	req := NewRootMessage(userUID)
	req.Payload = &pb.Message_ReqRoomList{
		ReqRoomList: &pb.ReqRoomList{},
	}

	stream.Send(req)
	log.Println("ReqRoomList client send :", req)
}

func ResRoomList(data *pb.Message) bool {

	res := data.GetResRoomList()
	if res == nil {
		log.Println("fail, GetResRoomList()")
		return false
	}

	log.Println("GetResRoomList server return : user id : ", data.UserID)

	for _, roomID := range res.RoomIDs {
		log.Println("GetResRoomList server return : room id : ", roomID)
	}

	return true
}

func ReqCreate(stream pb.Gateway_RouteMessageClient, userUID int64) {
	req := NewRootMessage(userUID)
	req.Payload = &pb.Message_ReqCreate{
		ReqCreate: &pb.ReqCreate{},
	}

	stream.Send(req)

	log.Println("ReqCreate client send :", req)
}

func ResCreate(data *pb.Message) bool {

	res := data.GetResCreate()
	if res == nil {
		log.Println("fail, GetReqLogin()")
		return false
	}

	log.Println("ResCreate server return : user id : ", data.UserID)
	log.Println("ResCreate server return : room id : ", res.RoomID)

	return true
}

func ReqJoin(stream pb.Gateway_RouteMessageClient, userUID int64, roomID int64) {
	req := NewRootMessage(userUID)
	req.Payload = &pb.Message_ReqJoin{
		ReqJoin: &pb.ReqJoin{
			RoomID: roomID,
		},
	}

	stream.Send(req)
	log.Println("ReqJoin client send :", req)
}

func ResJoin(data *pb.Message) bool {

	res := data.GetResJoin()
	if res == nil {
		log.Println("fail, GetReqLogin()")
		return false
	}

	log.Println("GetResJoin server return : user id : ", data.UserID)
	log.Println("GetResJoin server return : room id : ", res.RoomID)

	for _, memberID := range res.Members {
		log.Println("GetResJoin server return : member id : ", memberID)
	}

	return true
}

func ReqAction1(stream pb.Gateway_RouteMessageClient, userUID int64) {

	req := NewRootMessage(userUID)
	req.Payload = &pb.Message_ReqAction1{
		ReqAction1: &pb.ReqAction1{},
	}

	stream.Send(req)

	log.Println("ReqAction1 client send :", req)
}

func ResAction1(data *pb.Message) bool {

	res := data.GetResAction1()
	if res == nil {
		log.Println("fail, GetResAction1()")
		return false
	}

	log.Println("GetResAction1 server return : user id : ", data.UserID)
	log.Println("GetResAction1 server return : result : ", res.Result)

	return true
}

func ReqQuit(stream pb.Gateway_RouteMessageClient, userUID int64) {
	req := NewRootMessage(userUID)
	req.Payload = &pb.Message_ReqQuit{
		ReqQuit: &pb.ReqQuit{},
	}

	stream.Send(req)

	log.Println("ReqQuit client send :", req)
}

func ResQuit(data *pb.Message) bool {
	res := data.GetResQuit()
	if res == nil {
		log.Println("fail, GetResAction1()")
		return false
	}

	log.Println("GetResQuit server return : user id : ", res.IsSuccess)
	return true
}

func NotifyJoinHandler(data *pb.Message) bool {
	res := data.GetNotifyJoin()
	if res == nil {
		log.Println("fail, GetResAction1()")
		return false
	}

	log.Println("GetNotifyJoin server return : user id : ", data.UserID)
	log.Println("GetNotifyJoin server return : room id : ", res.RoomID)
	return true
}

func NotifyAction1Handler(data *pb.Message) bool {
	res := data.GetNotifyAction1()
	if res == nil {
		log.Println("fail, GetResAction1()")
		return false
	}

	log.Println("GetNotifyAction1 server return : user id : ", data.UserID)
	return true
}

func NotifyQuitHandler(data *pb.Message) bool {
	res := data.GetNotifyQuit()
	if res == nil {
		log.Println("fail, GetResAction1()")
		return false
	}

	log.Println("GetNotifyQuit server return : user id : ", data.UserID)
	log.Println("GetNotifyQuit server return : room id : ", res.RoomID)
	return true
}

func RunUserIdDialog(owner walk.Form) (int, error) {
	var dlg *walk.Dialog
	var acceptPB, cancelPB *walk.PushButton
	var inDlg *walk.LineEdit

	return walk_dcl.Dialog{
		AssignTo:      &dlg,
		Title:         "input User ID",
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		MinSize: walk_dcl.Size{
			Width: 200, Height: 100},
		Layout: walk_dcl.VBox{},
		Children: []walk_dcl.Widget{
			walk_dcl.Composite{
				Layout: walk_dcl.Grid{Columns: 2},
				Children: []walk_dcl.Widget{
					walk_dcl.Label{
						Text: "User ID:",
					},
					walk_dcl.LineEdit{
						AssignTo: &inDlg,
						Text:     "",
					},
				},
			},
			walk_dcl.Composite{
				Layout: walk_dcl.HBox{},
				Children: []walk_dcl.Widget{
					walk_dcl.HSpacer{},
					walk_dcl.PushButton{
						AssignTo: &acceptPB,
						Text:     "OK",
						OnClicked: func() {
							inputString = inDlg.Text()
							dlg.Accept()
						},
					},
					walk_dcl.PushButton{
						AssignTo: &cancelPB,
						Text:     "Cancel",
						OnClicked: func() {
							dlg.Cancel()
						},
					},
				},
			},
		},
	}.Run(owner)
}

func RunRoomJoinDialog(owner walk.Form) (int, error) {
	var dlg *walk.Dialog
	var acceptPB, cancelPB *walk.PushButton
	var inDlg *walk.LineEdit

	return walk_dcl.Dialog{
		AssignTo:      &dlg,
		Title:         "input Room ID",
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		MinSize:       walk_dcl.Size{Width: 200, Height: 100},
		Layout:        walk_dcl.VBox{},
		Children: []walk_dcl.Widget{
			walk_dcl.Composite{
				Layout: walk_dcl.Grid{Columns: 2},
				Children: []walk_dcl.Widget{
					walk_dcl.Label{
						Text: "room id:",
					},
					walk_dcl.LineEdit{
						AssignTo: &inDlg,
						Text:     "",
					},
				},
			},
			walk_dcl.Composite{
				Layout: walk_dcl.HBox{},
				Children: []walk_dcl.Widget{
					walk_dcl.HSpacer{},
					walk_dcl.PushButton{
						AssignTo: &acceptPB,
						Text:     "OK",
						OnClicked: func() {
							inputString = inDlg.Text()
							dlg.Accept()
						},
					},
					walk_dcl.PushButton{
						AssignTo: &cancelPB,
						Text:     "Cancel",
						OnClicked: func() {
							dlg.Cancel()
						},
					},
				},
			},
		},
	}.Run(owner)
}
