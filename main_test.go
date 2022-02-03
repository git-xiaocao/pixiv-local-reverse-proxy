package PixivLocalReverseProxy

import "testing"

func TestStartServer(t *testing.T) {
	isDebugMode = true
	enableLog = true
	StartServer("12345")
	//注意 如果你不手动关闭它 它就永远不会停止
}
