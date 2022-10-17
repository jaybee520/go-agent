package mq

import (
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"go-agent/src/cn/com/afcat/vo"
)

type MqConf struct {
	NameServers  []string `mapstructure:"nameServers"`
	InstanceName string
	Tag          string
}

var (
	MqProducer            rocketmq.Producer
	MqPushConsumerSuccess rocketmq.PushConsumer
	MqServer              vo.MqServer
)

const (
	MqRetryTimes = 3
)

func InitMq(mqConf *MqConf, server vo.MqServer) {
	if mqConf == nil {
		panic("mq config is nil")
		return
	}
	MqServer = server
	var err error
	MqProducer, err = rocketmq.NewProducer(
		producer.WithGroupName("notify_producer"),
		producer.WithNameServer(mqConf.NameServers),
		producer.WithRetry(MqRetryTimes),
	)
	if err != nil {
		panic(fmt.Sprintf("init rocket mq producer err:%v", err))
		return
	}

	err = MqProducer.Start()
	if err != nil {
		panic(fmt.Sprintf("producer mq start err:%v", err))
		return
	}

	MqPushConsumerSuccess, err = rocketmq.NewPushConsumer(
		consumer.WithGroupName("notify_consumer_success"),
		consumer.WithNameServer(mqConf.NameServers),
		consumer.WithInstance(mqConf.InstanceName),
		consumer.WithConsumerModel(consumer.BroadCasting),
	)
	if err != nil {
		panic(fmt.Sprintf("init rocket mq push consumer err:%v", err))
		return
	}
}

func ShutDownMq() {
	_ = MqProducer.Shutdown()
	_ = MqPushConsumerSuccess.Shutdown()
}
