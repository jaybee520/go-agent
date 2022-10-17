package afcat

import (
	"context"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"log"
	"os"
)

func Listener1(group string, addr string, instanceName string, topic string, tag string) {
	c, _ := rocketmq.NewPushConsumer(
		consumer.WithGroupName(group),
		consumer.WithNsResolver(primitive.NewPassthroughResolver([]string{addr})),
		consumer.WithInstance(instanceName),
		consumer.WithConsumerModel(consumer.BroadCasting),
	)

	err := c.Subscribe(topic, consumer.MessageSelector{Type: consumer.TAG, Expression: tag}, func(ctx context.Context,
		msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for i := range msgs {
			log.Printf("订阅到消息: body=%v, tag =%v \n", string(msgs[i].Body), msgs[i].GetTags())
		}
		return consumer.ConsumeSuccess, nil
	})
	if err != nil {
		log.Printf(err.Error())
	}
	err = c.Start()
	defer c.Shutdown()
	if err != nil {
		log.Printf(err.Error())
		os.Exit(-1)
	}
	//time.Sleep(20 * time.Minute)
}
