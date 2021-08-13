package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"logkv/protocol"
	"os"
	"strings"

	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
	_ "github.com/davyxu/cellnet/peer/tcp"
	"github.com/davyxu/cellnet/proc"
	_ "github.com/davyxu/cellnet/proc/tcp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	var ctx, cancel = context.WithCancel(context.Background())
	// 创建一个事件处理队列，整个客户端只有这一个队列处理事件，客户端属于单线程模型
	queue := cellnet.NewEventQueue()

	// 创建一个tcp的连接器，名称为client，连接地址为127.0.0.1:8801，将事件投递到queue队列,单线程的处理（收发封包过程是多线程）
	p := peer.NewGenericPeer("tcp.Connector", "client", "127.0.0.1:3210", queue)

	// 设定封包收发处理的模式为tcp的ltv(Length-Type-Value), Length为封包大小，Type为消息ID，Value为消息内容
	// 并使用switch处理收到的消息
	proc.BindProcessorHandler(p, "tcp.ltv", func(ev cellnet.Event) {
		switch msg := ev.Message().(type) {
		case *cellnet.SessionConnected:
			log.Println("client connected")
		case *cellnet.SessionClosed:
			cancel()
			log.Println("client error")
			return
		case *cellnet.SessionConnectError:
			cancel()
			return
		case *protocol.SetAck:
			log.Println(msg)
		case *protocol.GetAck:
			fmt.Printf("%d:%s,%s\n", msg.Code, msg.Message, msg.Data)
			var v = map[string]interface{}{}
			err := bson.Unmarshal(msg.Data, &v)
			fmt.Println(v, err)
		case *protocol.BatchGetAck:
		case *protocol.BatchSetAck:
		case *protocol.DeleteAck:
		case *protocol.ScanAck:
		default:
			log.Println(msg)
		}
	})

	// 开始发起到服务器的连接
	p.Start()

	// 事件队列开始循环
	queue.StartLoop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				os.Exit(0)
				return
			}
		}
	}()

	// 阻塞的从命令行获取聊天输入
	ReadConsole(ctx, func(str string) {

		s := strings.Split(str, " ")
		if len(s) <= 1 {
			log.Println("unkown cmd")
			return
		}
		var sess = p.(interface {
			Session() cellnet.Session
		}).Session()
		switch s[0] {
		case "set":
			key := primitive.NewObjectID()
			var v = Log{
				Id:     key,
				App:    "main",
				Custom: strings.Join(s[1:], " "),
			}
			data, err := bson.Marshal(v)
			if err != nil {
				log.Println(err)
				return
			}
			var req = protocol.SetReq{
				Data: data,
			}
			sess.Send(req)
			log.Println("set", key.Hex())
		case "get":
			var key, err = primitive.ObjectIDFromHex(s[1])
			if err != nil {
				log.Println(err)
				return
			}

			var req = protocol.GetReq{
				Key: key.Hex(),
			}

			sess.Send(&req)
		default:
			log.Println("unkown cmd", str)
		}

	})
}

func ReadCmd() {

}

func ReadConsole(ctx context.Context, callback func(string)) {

	for {

		select {
		case <-ctx.Done():
			return
		default:
		}

		// 从标准输入读取字符串，以\n为分割
		text, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			break
		}

		// 去掉读入内容的空白符
		text = strings.TrimSpace(text)

		callback(text)

	}

}
