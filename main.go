package main

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/grpc-playground/proto/pb"

	"google.golang.org/grpc/codes"

	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

var (
	rootCtx = context.Background()
)

func main() {
	// 连接服务器
	ctx, cancel := context.WithTimeout(rootCtx, time.Second*5)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		"127.0.0.1:50051",
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("faild to connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewGreeterClient(conn)

	log.Println("\n\n\nTest Bidirectional streaming")
	sayHello(c)
}

func sayHello(c pb.GreeterClient) {
	var err error

	stream, err := c.SayHello(context.Background())
	if err != nil {
		log.Printf("failed to call: %v", err)
		return
	}

	var i int64
MainLoop:
	for {
		stream.Send(&pb.HelloRequest{Name: "name " + strconv.FormatInt(i, 10)})
		if err != nil {
			log.Printf("failed to send: %v", err)
			break
		}
		reply, err := stream.Recv()
		if err != nil {
			log.Printf("failed to recv: %v", err)
			statusError, ok := status.FromError(err)
			if ok {
				if statusError.Code() == codes.Unavailable {
					for {
						log.Println("服务器意外退出 正在重连中...")
						stream, err = c.SayHello(rootCtx)
						if err != nil {
							log.Printf("failed to call: %v", err)
							time.Sleep(time.Second * 2)
							continue
						}
						goto MainLoop
					}
				}
			}
			break
		}
		log.Printf("Greeting: %s", reply.Message)
		i++
		time.Sleep(time.Second * 1)
	}
}
