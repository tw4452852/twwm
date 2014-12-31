package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc"

	"github.com/tw4452852/twwm"
)

func init() {
	for _, r := range twwm.Recevers {
		rpc.Register(r)
	}
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":54321")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}
