package main

import (
	"flag"
	"log"
)

func maybe(err error) {
	if err != nil {
		log.Fatalf("error: %s", err.Error())
	}
}

func main() {
	cmd := flag.String("cmd", "", "[server,client]")
	flag.Parse()
	switch *cmd {
	case "server":
		runServer()
	case "client":
	default:
		log.Fatalf("bad command: %s", *cmd)
	}
}
