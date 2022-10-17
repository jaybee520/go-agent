package logService

import (
	"encoding/binary"
	"github.com/go-netty/go-netty"
	"github.com/go-netty/go-netty/codec/frame"
	"go-agent/src/cn/com/afcat/format"
	"log"
	"time"
)

var (
	//LogClient *channel.Channel
	logClient netty.Channel
	addRess   string
)

type LogConnConf struct {
	address string
}

func InitLogClient(address string) {
	addRess = address
	//conn, err := net.Dial("tcp", address)
	//if err != nil {
	//	fmt.Println("连接服务器错误")
	//	return
	//}
	//
	//LogClient = channel.NewChannel(conn)

	/*childInitializer := func(channel netty.Channel) {
		channel.Pipeline().
			//AddLast(frame.LengthFieldCodec(binary.LittleEndian, 1024, 0, 2, 0, 2)).
			AddLast(format.TextCodec())
		//AddLast(EchoHandler{"Server"})
	}*/

	// setup client pipeline initializer.
	clientInitializer := func(channel netty.Channel) {
		channel.Pipeline().
			AddLast(frame.LengthFieldCodec(binary.BigEndian, 1048576, 0, 4, 0, 4)).
			AddLast(format.ObjectCodec())
		//AddLast(EchoHandler{"Client"})
	}

	// new bootstrap
	var bootstrap = netty.NewBootstrap(netty.WithClientInitializer(clientInitializer))
	connect, err := bootstrap.Connect(address, nil)
	if err != nil {
		log.Println("日志服务器连接失败：", err)
		panic(err)
	}
	logClient = connect
}
func GetLogClient() netty.Channel {
	if logClient != nil && logClient.IsActive() {
		return logClient
	}
	defer func() {
		if err := recover(); err != nil {
			log.Println("消息服务器连接失败3秒后重新连接，%v", err)
			time.Sleep(time.Second * 3)
			InitLogClient(addRess)
		}
	}()
	InitLogClient(addRess)
	return logClient
}

func SenHeart() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("日志服务器心跳发送失败，%v", err)
			time.Sleep(time.Second)
			SenHeart()
		}
	}()
	for {
		GetLogClient().Write("HEART_BEAT_TASK")
		time.Sleep(3 * time.Second)
	}
}
