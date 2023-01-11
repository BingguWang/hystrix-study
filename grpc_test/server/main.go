package main

import (
	"flag"
	pb "github.com/BingguWang/hystrix-study/grpc_test/server/proto"
	"github.com/BingguWang/hystrix-study/grpc_test/server/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
)

var (
	host = flag.String("host", "localhost", "")
	port = flag.String("port", "50051", "")
)

func main() {
	flag.Parsed()
	addr := net.JoinHostPort(*host, *port)
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	server := grpc.NewServer(
		grpc.Creds(insecure.NewCredentials()),
	)

	pb.RegisterScoreServiceServer(server, &service.ServiceImpl{})

	if err := server.Serve(listen); err != nil {
		return
	}
}
