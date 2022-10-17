package logService

import (
	"encoding/json"
	"go-agent/src/cn/com/afcat/vo"
)

func SendLog(vo vo.LogMessageVO) {
	marshal, _ := json.Marshal(vo)

	//log.Println("发送日志: %v", string(marshal))
	GetLogClient().Write(string(marshal))
	//conf.LogClient.Write("HEART_BEAT_TASK")
}
