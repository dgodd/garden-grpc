package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os/exec"

	pb "github.com/dgodd/garden-grpc/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 0, "The server port")
)

type containerServer struct {
	done chan int
}

func newServer() *containerServer {
	s := new(containerServer)
	s.done = make(chan int)
	return s
}

func (c *containerServer) Run(req *pb.RunRequest, stdout pb.Container_RunServer) error {
	log.Println("Hi -- doing run:", req.Path)
	cmd := exec.Command(req.Path)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}
	fmt.Println("out:", out)
	// go io.Copy(os.Stdout, out)

	r := bufio.NewReader(out)
	line, isPrefix, err := r.ReadLine()
	for err == nil && !isPrefix {
		stdout.Send(&pb.Stdout{Line: string(line)})
		line, isPrefix, err = r.ReadLine()
	}

	cmd.Wait()
	return nil
}

func (c *containerServer) Exit(ctx context.Context, req *pb.ExitRequest) (*pb.ExitRequest, error) {
	// os.Exit(int(req.Code))
	c.done <- 4
	return req, nil
}

func main() {
	defer func() { fmt.Println("AtExit do stuff") }()

	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println("Listening on :", lis.Addr().(*net.TCPAddr).Port)
	grpcServer := grpc.NewServer()
	containerServer := newServer()
	pb.RegisterContainerServer(grpcServer, containerServer)
	go grpcServer.Serve(lis)

	<-containerServer.done
}
