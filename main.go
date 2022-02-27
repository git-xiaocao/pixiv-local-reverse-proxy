package PixivLocalReverseProxy

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"github.com/elazarl/goproxy"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	isDebugMode = false
	dnsCache    = struct {
		Data map[string]string
		Lock sync.RWMutex
	}{
		Data: make(map[string]string),
		Lock: sync.RWMutex{},
	}
	tlsConfigCache = struct {
		Data map[string]*tls.Config
		Lock sync.RWMutex
	}{
		Data: make(map[string]*tls.Config),
		Lock: sync.RWMutex{},
	}
	delay      = 10 * time.Second
	pixivHosts = []string{
		"pixiv.net",
		"www.pixiv.net",
		"app-api.pixiv.net",
		"oauth.secure.pixiv.net",
		"source.pixiv.net",
		"accounts.pixiv.net",
		"touch.pixiv.net",
		"imgaz.pixiv.net",
		"dic.pixiv.net",
		"comic.pixiv.net",
		"factory.pixiv.net",
		"g-client-proxy.pixiv.net",
		"sketch.pixiv.net",
		"payment.pixiv.net",
		"sensei.pixiv.net",
		"novel.pixiv.net",
		"en-dic.pixiv.net",
		"i1.pixiv.net",
		"i2.pixiv.net",
		"i3.pixiv.net",
		"i4.pixiv.net",
		"d.pixiv.org",
		"fanbox.pixiv.net",
		"pixivsketch.net",
		"pximg.net",
		"i.pximg.net",
		"s.pximg.net",
		"pixiv.pximg.net",
	}
	localCa tls.Certificate
	server  *http.Server
)

func pixivConnectHijack(req *http.Request, conn net.Conn, ctx *goproxy.ProxyCtx) {
	hostname := req.URL.Hostname()
	port := ctx.Req.URL.Port()

	defer func() {
		if r := recover(); r != nil {
			_, _ = conn.Write([]byte("HTTP/1.1 500\r\n\r\n"))
		}
		_ = conn.Close()
	}()
	tlsConfigCache.Lock.RLock()
	tlsConfig, hasTlsConfig := tlsConfigCache.Data[hostname]
	tlsConfigCache.Lock.RUnlock()
	if !hasTlsConfig {
		var err error
		tlsConfig, err = goproxy.TLSConfigFromCA(&localCa)(hostname, ctx)
		if err != nil {
			panic(err)
		} else {
			tlsConfigCache.Lock.Lock()
			tlsConfigCache.Data[hostname] = tlsConfig
			tlsConfigCache.Lock.Unlock()
		}
	}

	tlsConn := tls.Server(conn, tlsConfig)

	defer func() {
		_ = tlsConn.Close()
	}()

	_ = tlsConn.SetDeadline(time.Now().Add(delay))

	clientIO := bufio.NewReadWriter(bufio.NewReader(tlsConn), bufio.NewWriter(tlsConn))

	remoteCon := buildNewDnsConn(hostname, port)
	if remoteCon == nil {
		panic("创建新的连接失败:" + hostname)
	}

	defer func() {
		_ = remoteCon.Close()
	}()

	remote := tls.Client(remoteCon, &tls.Config{InsecureSkipVerify: true})

	defer func() {
		_ = remote.Close()
	}()

	remoteIO := bufio.NewReadWriter(bufio.NewReader(remote), bufio.NewWriter(remote))

	errChan := make(chan error)

	go func() {
		_, _ = conn.Write([]byte("HTTP/1.1 200\r\n\r\n"))
		var buffer = make([]byte, 10240)
		var err error
		for {
			n, err := clientIO.Read(buffer)
			if err != nil {
				break
			}
			_, err = remoteIO.Write(buffer[:n])
			if err != nil {
				break
			}
			if err := remoteIO.Flush(); err != nil {
				break
			}
		}
		errChan <- err
	}()

	go func() {
		var buffer = make([]byte, 10240)
		var err error
		for {
			n, err := remoteIO.Read(buffer)
			if err != nil {
				break
			}
			_, err = clientIO.Write(buffer[:n])
			if err != nil {
				break
			}
			if err := clientIO.Flush(); err != nil {
				break
			}
		}
		errChan <- err
	}()

	if err := <-errChan; err != nil {
		panic(err)
	}

}

func buildNewDnsConn(hostname, port string) net.Conn {
	dnsCache.Lock.RLock()

	data, ok := dnsCache.Data[hostname]

	dnsCache.Lock.RUnlock()
	//在缓存里
	if ok {
		remoteCon, err := net.Dial("tcp", data+":"+port)
		if err != nil {
			return nil
		}
		return remoteCon
	}

	DnsQueryRequest := dnsQueryRequest{
		hostname,
	}

	res, err := DnsQueryRequest.fetch()
	if err != nil {
		panic(err)
	}

	for _, answer := range res.Answer {
		if answer.Type != 1 {
			continue
		}
		remoteCon, _ := net.Dial("tcp", answer.Data+":"+port)
		dnsCache.Lock.Lock()
		dnsCache.Data[hostname] = answer.Data
		dnsCache.Lock.Unlock()
		return remoteCon
	}
	return nil
}

func StartServer(bindPort string, enableLog bool) {
	localCa, _ = tls.X509KeyPair(caCert, caKey)
	localCa.Leaf, _ = x509.ParseCertificate(localCa.Certificate[0])
	hijacks := make([]string, len(pixivHosts))

	for i, s := range pixivHosts {
		hijacks[i] = s + ":443"
	}

	proxy := goproxy.NewProxyHttpServer()

	proxy.OnRequest(
		goproxy.ReqHostIs(hijacks...),
	).HijackConnect(pixivConnectHijack)

	proxy.OnRequest().HandleConnect(goproxy.AlwaysReject)

	proxy.Verbose = enableLog

	server = &http.Server{Addr: ":" + bindPort, Handler: proxy}

	if isDebugMode {
		_ = server.ListenAndServe()
	} else {
		go func() {
			_ = server.ListenAndServe()
		}()
	}
}

func StopServer() {
	if server != nil {
		_ = server.Shutdown(context.Background())
		_ = server.Close()
		server = nil
	}
}
