package main

import (
	"main/logs"

	_ "github.com/lib/pq"
)

func main() {

	// init the logger
	_ = logs.InitLogger()

	// mux
	mux := Routes()
	server := NewServer(mux)
	server.Run()

}
