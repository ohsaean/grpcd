package main

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"github.com/ohsaean/grpcd/lib"
	pb "github.com/ohsaean/grpcd/proto"
	"io"
	"log"
	"net"
)

var (
	rooms lib.SharedMap
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

	lib.Log("RouteMessage 호출됨")
	go onClientRead(stream, user)

	// send loop 처리
	for {
		select {
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
}

func main() {
	InitRooms()
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGatewayServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
