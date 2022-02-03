package main

import "C"
import (
	PixivLocalReverseProxy "pixiv-local-reverse-proxy"
	"strconv"
)

//export StartServer
func StartServer(bindPort uint16) {
	PixivLocalReverseProxy.StartServer(strconv.Itoa(int(bindPort)))
}

//export StopServer
func StopServer() {

}

func main() {

}
