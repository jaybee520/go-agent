package listener

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go-agent/src/cn/com/afcat/mq"
	"go-agent/src/cn/com/afcat/redis"
	"go-agent/src/cn/com/afcat/task"
	"go-agent/src/cn/com/afcat/vo"
	"log"
)

type MqConsumer struct {
	Topic string
	Tag   string
}

func SubScribe(mqConsumer MqConsumer) {
	var err error

	err = mq.MqPushConsumerSuccess.Subscribe(mqConsumer.Topic, consumer.MessageSelector{Type: consumer.TAG, Expression: mqConsumer.Tag}, callBackSucc)
	if err != nil {
		log.Println("consumer subscribe TopicNotifySucess err:%v", err)
		return
	}

	err = mq.MqPushConsumerSuccess.Start()
	if err != nil {
		log.Println("MqPushConsumerSuccess start err:%v", err)
		return
	}

}
func callBackSucc(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("消息处理失败:%s", err)
			log.Println("消息处理失败:%s", msgs)
		}
	}()
	for i := range msgs {
		msgString := string(msgs[i].Body)
		fmt.Println(msgString)
		var messageVO vo.MessageVO
		err := json.Unmarshal(msgs[i].Body, &messageVO)
		if err != nil {
			log.Println("消息解析失败: %s" + msgString)
			return consumer.ConsumeRetryLater, nil
		}
		defer func() {
			if err := recover(); err != nil {
				log.Println("消息处理失败:%s", err)
				log.Println("消息处理失败:%s", msgString)
			}
		}()
		if redis.Exists(messageVO.Key) {
			redis.Del(messageVO.Key)
			go task.TaskExeService(messageVO)
		} else {
			log.Println("消息KEY不是有效数据: %s" + msgString)
		}
	}
	return consumer.ConsumeSuccess, nil
}
