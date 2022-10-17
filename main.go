package main

import (
	"snakeAndLadder/core"
	"snakeAndLadder/edge/http"
	"snakeAndLadder/edge/ws"
)

var (
	localhost = "0.0.0.0:8080"
)

func main() {
	/*
		init ws
	*/
	initWebsocket()
	/*
		init grpc svr
	*/
	core.RunBackendServer()
	/*
		initEcho
	*/
	e := http.InitRouter()
	e.Logger.Fatal(e.Start(localhost))
}

func initWebsocket() {
	var err error
	ws.WebsocketServer, err = ws.NewWebsocketServer()
	if err != nil {
		panic("init wsSvr fail, err:" + err.Error())
	}
}
