package main

import (
	"bufio"
	"log"
	"logkv/protocol"
	"os"
	"strings"

	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/peer"
	_ "github.com/davyxu/cellnet/peer/tcp"
	"github.com/davyxu/cellnet/proc"
	_ "github.com/davyxu/cellnet/proc/tcp"
	"gopkg.in/mgo.v2/bson"
)

func main() {
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
			log.Println("client error")
		case *protocol.SetAck:
			log.Println(msg)
		case *protocol.GetAck:
			log.Println(protocol.KvFromBytes(msg.Data))
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

	log.Println("connected")

	// 阻塞的从命令行获取聊天输入
	ReadConsole(func(str string) {

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
			key := bson.NewObjectId()
			var req = protocol.SetReq{
				Key:  key,
				Data: []byte(s[1]),
			}
			sess.Send(req)
			log.Println("set", key.Hex())
		case "get":
			key := bson.ObjectIdHex(s[1])
			var req = protocol.GetReq{
				Key: key,
			}
			sess.Send(&req)
		default:
			log.Println("unkown cmd")
		}

	})
}

func ReadConsole(callback func(string)) {

	for {

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
