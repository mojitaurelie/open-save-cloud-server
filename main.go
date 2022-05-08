package main

import (
	"fmt"
	"opensavecloudserver/constant"
	"opensavecloudserver/server"
	"runtime"
)

func main() {
	fmt.Printf("Open Save Cloud (Server) %s (%s %s)\n", constant.Version, runtime.GOOS, runtime.GOARCH)
	server.Serve()
}
