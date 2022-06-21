package main

import (
	"crypto/tls"
	"flag"
	"log"

	example "github.com/rpcx-ecosystem/rpcx-examples3"
	"github.com/smallnest/rpcx/server"
)

var (
	addr = flag.String("addr", "localhost:8888", "server address")
)

func main() {
	flag.Parse()

	// 读取私钥/公钥文件并进行处理
	cert, err := tls.LoadX509KeyPair("server.pem", "server.key")
	if err != nil {
		log.Println(err)
	}

	config := &tls.Config{Certificates: []tls.Certificate{cert}}

	// 注册Airth
	s := server.NewServer(server.WithTLSConfig(config))
	s.RegisterName("Arith", new(example.Arith), "")

	// 启动一个quic服务
	err = s.Serve("quic", *addr)
	if err != nil {
		panic(err)
	}
}
