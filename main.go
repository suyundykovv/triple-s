package main

import (
	"triple-s/flags"
	server "triple-s/servers"
)

func main() {
	flags.Setup()
	server.Start(flags.Port)
}
