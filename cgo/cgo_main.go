package main

import "C"
import (
	PixivLocalReverseProxy "pixiv-local-reverse-proxy"
	"strconv"
)

//export StartServer
func StartServer(bindPort uint16, enableLog bool) {
	PixivLocalReverseProxy.StartServer(strconv.Itoa(int(bindPort)), enableLog)
}

//export StopServer
func StopServer() {
	PixivLocalReverseProxy.StopServer()
}

func main() {

}
