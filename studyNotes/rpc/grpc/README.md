# GRPC


## 概念
### gRPC
gRPC是Google开源的RPC框架和库,已支持主流计算机语言。底层通信采用 gRPC 协议，比较适合互联网场景。gRPC 在设计上考虑了跟 ProtoBuf 的配合使用。

### ProtoBuf
ProtoBuf是一套接口描述语言(IDL)和工具集(主要是protoc, 基于C++实现),类似Apache的thrift。用户写好.proto描述文件,之后使用protoc可以很容易编译成众多计算机语言(C++、Java、Python、C#、Golang等)的接口代码。这些代码可以支持gRPC,也可以不支持。


### 使用场景
写好.proto描述文件定义RPC的接口,然后用protoc(带gRPC插件)的.proto模版自动生成客户端和服务端的接口代码。


## Protobuf

### 工具集

* 编译器: protoc,以及一些官方没有带的语言插件。
* 运行环境: 各种语言的protobuf库,不同语言有不同的安装来源。

Protobuf有一些自己的语言规范。
 
 * message: 代表数据结构,可以包括不同类型的成员变量。包括字符串、数字、数组、字典等。变量后面的数字代表进行二进制编码时候的提示信息, 1~15表示热变量(1Byte)。默认所有变量都是可选的(optional)，repeated则表示数组。
 * service: 代表RPC接口。
  
 ```
 syntax = "proto3";
 package hello;
  
 message HelloRequest {
 	string greeting = 1; 
 }
 
 message HelloResponse {
 	string reply = 1;
 	repeated int32 number = 4;
 }
 
 service HelloService {
 	rpc SayHello(HelloRequest) returns (HelloResponse){} 
 }
 ```
 
### 编译
编译最关键参数是制定输出语言格式。一些没有官方支持的语言,可以通过安装protoc对应的plugin来支持。

```
// Go
$ go get -u github.com/golang/protobuf/{protoc-gen-go,proto}
```

一些准备完之后,就可以进行编译。

```
// 生成hello.pb.go,会自动调用protoc-gen-go插件。
$ protoc --go_out=./ hello.proto
```

## gRPC
gRPC的库在服务端提供一个gPRC server,客户端提供一个gRPC Stub。典型的应用场景是客户端发送请求(同步或异步)访问服务端的接口。客户端和服务端之间的通信协议是基于HTTP2的gRPC协议,支持全双工的流式保序消息,性能比较好,同时也很轻。

采用ProtoBuf作为IDL,则需要定义service类型。生成客户端和服务端代码,用户自行实现服务端代码中的调用接口,并且利用客户端代码来发起请求道服务端。



 
 

 
 
 
 