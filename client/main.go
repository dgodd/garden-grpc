package main

import (
	"flag"
	"fmt"
	"io"
	"log"

	pb "github.com/dgodd/garden-grpc/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	serverAddr = flag.String("server_addr", "127.0.0.1:10000", "The server address in the format of host:port")
)

func doRun(client pb.ContainerClient) {
	stream, err := client.Run(context.Background(), &pb.RunRequest{Path: "/sbin/ifconfig"})
	if err != nil {
		log.Fatalf("%v.Run(_) = _, %v", client, err)
	}
	for {
		stdout, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.Run(_) = _, %v", client, err)
		}
		fmt.Println(stdout.Line)
	}
}

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewContainerClient(conn)

	doRun(client)

	client.Exit(context.Background(), &pb.ExitRequest{Code: 3})
}
