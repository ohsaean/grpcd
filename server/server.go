package main

import (
	"github.com/ohsaean/grpcd/lib"
	pb "github.com/ohsaean/grpcd/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"io"
	"net"
)

var (
	rooms    lib.SharedMap
	sessions lib.SharedMap
)

type server struct{}

func onClientRead(stream pb.Gateway_RouteMessageServer, user *User) {
	// read loop 처리
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			return
		}

		switch in.Payload.(type) {

		case *pb.Message_ReqLogin:
			loginHandler(in, user)

		case *pb.Message_ReqCreate:
			createHandler(in, user)

		case *pb.Message_ReqJoin:
			joinHandler(in, user)

		case *pb.Message_ReqAction1:
			action1Handler(in, user)

		case *pb.Message_ReqRoomList:
			roomListHandler(in, user)

		case *pb.Message_ReqQuit:
			quitHandler(in, user)

		default:
			lib.Log("failed, not defined handler")
		}
	}
}

func (s *server) RouteMessage(stream pb.Gateway_RouteMessageServer) (err error) {

	user := NewUser(0, nil) // empty user data
	go onClientRead(stream, user)
	defer user.Leave()
	// send loop 처리
	for {
		select {
		case <-user.exit:
			// when receive signal then finish the program

			lib.Log("Leave user id :" + lib.Itoa64(user.userID))

			return
		case message := <-user.recv:
			stream.Send(message)
		}
	}
	return nil
}

const (
	port = ":50051"
)

func InitRooms() {
	rooms = lib.NewSMap(lib.RWMutex)
	sessions = lib.NewSMap(lib.RWMutex)
}

func main() {
	lib.Log("server start!")
	InitRooms()
	lis, err := net.Listen("tcp", port)
	if err != nil {
		lib.Logf("failed to listen: %v", err)
		return
	}
	s := grpc.NewServer()
	pb.RegisterGatewayServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		lib.Logf("failed to serve: %v", err)
		return
	}
}
