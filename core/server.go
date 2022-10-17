package core

import (
	"google.golang.org/grpc"
	"log"
	"net"
	api "snakeAndLadder/api"
	"snakeAndLadder/core/snakeAndLadder"
)

func RunBackendServer() {
	svr, listener, err := initGrpcServer()
	if err != nil {
		return
	}
	go runGrpcServer(svr, listener)

	return
}

func initGrpcServer() (*grpc.Server, net.Listener, error) {
	listen, err := net.Listen("tcp", ":5050")
	if err != nil {
		log.Printf("listen fail, err:%v", err)
		return nil, nil, err
	}
	var opts []grpc.ServerOption
	// append codec、naming
	svr := grpc.NewServer(opts...)
	api.RegisterSnakeAndLadderServer(svr, snakeAndLadder.NewService())
	return svr, listen, nil
}

func runGrpcServer(svr *grpc.Server, listener net.Listener) {
	//go func() {
	//	for {
	//		select {
	//		case <-done:
	//			svr.GracefulStop()
	//		default:
	//
	//		}
	//	}
	//}()
	err := svr.Serve(listener)
	if err != nil {
		log.Printf("grpc serve fail, err:%v", err)
		return
	}
}
