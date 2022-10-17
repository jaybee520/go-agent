package mq

import (
	"context"
	"encoding/json"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go-agent/src/cn/com/afcat/redis"
	"go-agent/src/cn/com/afcat/vo"
	"strconv"
	"time"
)

const (
	RetryTime = 3
	SleepTime = time.Millisecond * 10
)

func SendMq(message vo.ReturnMessageVO) (err error) {
	millisecond := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	message.Key = MqServer.Tag + millisecond
	marshal, _ := json.Marshal(&message)
	redis.SetData(message.Key, []byte("1"))
	msg := primitive.NewMessage(MqServer.Topic, []byte(marshal))
	msg.WithTag(MqServer.Tag)
	for i := 0; i < RetryTime; i++ {
		_, err := MqProducer.SendSync(context.Background(), msg)
		if err != nil {
			time.Sleep(SleepTime)
			continue
		}
		break
	}
	return

}
