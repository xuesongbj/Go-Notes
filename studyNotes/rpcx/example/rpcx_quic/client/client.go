package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log"

	example "github.com/rpcx-ecosystem/rpcx-examples3"
	"github.com/smallnest/rpcx/client"
)

var (
	addr = flag.String("addr", "localhost:8888", "server address.")
)

func main() {
	flag.Parse()

	conf := &tls.Config{

		// 控制客户端是否验证服务器的证书链和主机名。
		// 如果InsecureSkipVerify为true，则TLS接受服务器提供的任何证书以及该证书中的任何主机名。
		// 建议用在测试环境.
		InsecureSkipVerify: true,
	}

	option := client.DefaultOption
	option.TLSConfig = conf

	d := client.NewPeer2PeerDiscovery("quic@"+*addr, "")
	xclient := client.NewXClient("Arith", client.Failtry, client.RandomSelect, d, option)
	defer xclient.Close()

	args := &example.Args{
		A: 10,
		B: 20,
	}

	reply := &example.Reply{}
	err := xclient.Call(context.Background(), "Mul", args, reply)
	if err != nil {
		log.Fatalf("failed to call: %v", err)
	}

	log.Printf("%d * %d = %d", args.A, args.B, reply.C)
}
