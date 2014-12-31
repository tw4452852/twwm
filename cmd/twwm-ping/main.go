package main

import (
	"log"
	"net/rpc"
)

func main() {
	c, e := rpc.DialHTTP("tcp", "127.0.0.1:54321")
	if e != nil {
		log.Fatal(e)
	}
	defer c.Close()

	e = c.Call("Ping.Do", struct{}{}, &struct{}{})
	if e != nil {
		log.Fatal(e)
	}
}
