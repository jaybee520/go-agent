package main

import (
	"bytes"
	"fmt"
	"go-agent/src/cn/com/afcat/listener"
	"go-agent/src/cn/com/afcat/logService"
	"go-agent/src/cn/com/afcat/mq"
	"go-agent/src/cn/com/afcat/prop"
	"go-agent/src/cn/com/afcat/redis"
	"go-agent/src/cn/com/afcat/task"
	"go-agent/src/cn/com/afcat/util"
	"go-agent/src/cn/com/afcat/vo"
	"log"
	"os"
	"reflect"
	"strconv"
	"time"
)

func main() {
	name, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	addr, _ := util.ExternalIP()
	prop.Ip = addr.String()
	log.Println("当前主机IP：", prop.Ip)
	prop.AgentId = prop.Ip + "@" + name
	// 日志
	logService.InitLogClient("10.24.24.11:9999")
	go logService.SenHeart()
	// redis
	redis.RedisConf([]string{"10.24.24.11:26379", "10.24.24.11:26389", "10.24.24.11:26399"}, "YWZjYXRAMjAyMQ==")
	// 消息服务器
	mqConf := mq.MqConf{NameServers: []string{"10.24.24.11:9876"}, InstanceName: name, Tag: "return"}
	mq.InitMq(&mqConf, vo.MqServer{Topic: "return_uyw_dev", Tag: "return"})
	listener.SubScribe(listener.MqConsumer{Topic: "return_uyw_dev", Tag: name})
	// 启动心跳定时任务
	go task.HeartCheack()
	select {}
}

func TestMessageExe() {
	/*jsonString := []byte(`{
	    "detailsId":"202207191517192700286",
	    "exeSeq":"202207191517192300285",
	    "key":"DockerAgent:d6079d7b-1c0a-4644-a7a0-97a1b7b57769",
	    "param":"{\"framework\":{\"ServerType\":\"Linux\",\"ExecServer_Passwd\":\"UVdaallYUkFNREl3TWc9PQ==\",\"ScriptType\":\"shell\",\"ExecuteWay\":\"REMOTE\",\"ExecServer_User\":\"afcat\",\"ScriptPro\":\"user-defined\",\"ExecServer_IP\":\"10.24.24.11\",\"FileDownloadUrl\":\"http://10.24.24.53:8012/file_service/resources/shell/2020-08-05-15/202001201403557400010_1596612866576_27.95664413316137.sh\",\"JobID\":\"e1d1d401-0731-11ed-bc8a-0242ac130006_202207191517192700286_202207191517192300285\"},\"business\":{\"OutputPara\":[{\"paramData\":\"NULL\",\"paramName\":\"c\",\"paramType\":\"string\",\"required\":\"0\"},{\"paramData\":\"NULL\",\"paramName\":\"d\",\"paramType\":\"string\",\"required\":\"0\"}],\"InputPara\":[{\"paramData\":\"1\",\"paramName\":\"e\",\"paramType\":\"string\",\"required\":\"0\"},{\"paramData\":\"2\",\"paramName\":\"f\",\"paramType\":\"string\",\"required\":\"0\"}],\"EnvPara\":[]},\"timeOut\":\"0\"}",
	    "processId":"e1d1d401-0731-11ed-bc8a-0242ac130006",
	    "system":"UYW",
	    "taskId":"sid-E34366F7-C2CE-49A8-9C38-583C924EDACF",
	    "taskName":"lin-test1",
	    "Type":"CMD",
	    "time":1658215504144
	}`)
	var messageVO vo.MessageVO
	json.Unmarshal(jsonString, &messageVO)
	//var paramVO vo.ParamVO
	//json.Unmarshal([]byte(messageVO.Param), &paramVO)
	//println(len(paramVO.Business.InputPara))
	task.TaskExeService(messageVO)
	print(messageVO.Param)*/

	//util.ExecCmd()
	//afcat.Listener1("testGroup", "10.24.24.53:9876", "abc", "return_uyw_dev", "abc")
	/*name, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	mqConf := conf.MqConf{NameServers: []string{"10.24.24.53:9876"}, InstanceName: name, Tag: "return"}
	conf.InitMq(&mqConf, vo.MqServer{Topic: "return_uyw_dev", Tag: "abc"})
	listener.SubScribe(listener.MqConsumer{Topic: "return_uyw_dev", Tag: "abc"})
	go send()*/
}

const STREAM_MAGIC = 0xACED
const STREAM_VERSION = 5

func redisRead() {
	redis.RedisConf([]string{"10.24.24.11:26379", "10.24.24.11:26389", "10.24.24.11:26399"}, "YWZjYXRAMjAyMQ==")

	jo := GetData()

	classes := jo.Classes
	fields := classes[0].Fields
	fmt.Println(fields[3].FieldValue)
	for _, v := range fields {
		if "time" == v.FieldName {
			v.FieldValue = time.Now().Format("2006-01-02 15:04:05")
		}
		if "version" == v.FieldName {
			v.FieldValue = "2.0"
		}
	}
	/*jsonMap := jo.JsonMap()
	s := jsonMap.(map[string]interface{})
	i := s["version"]
	fmt.Println(i)
	s["version"] = "2.0"*/
	timeStr := time.Now().Format("2006-01-02 15:04:05")
	buffer1 := new(bytes.Buffer)
	fmt.Println(timeStr)
	util.SerializeJavaEntity(buffer1, jo)
	redis.SetData("heart:host:127.0.0.1", buffer1.Bytes())
	GetData()
}

func GetData() *util.JavaTcObject {
	bytes1 := redis.GetData("heart:host:127.0.0.1")

	log.Println(string(bytes1))

	buffer := bytes.NewBuffer(bytes1)
	arr := make([]byte, 1<<7) //128

	if _, err := buffer.Read(arr[:4]); err != nil {
		//to be continued...
		log.Println("Got error %v\n", err)
	}
	refs := make([]*util.JavaReferenceObject, 1<<7)
	jo := &util.JavaTcObject{}
	if err := jo.Deserialize(buffer, refs); err != nil {
		log.Println("When deserialize JavaTcObject got %v\n", err)
	}
	classes := jo.Classes
	fields := classes[0].Fields

	fmt.Println(fields[3].FieldValue)
	log.Println("Got Tc_OBJECT %v\n", jo)
	rv := reflect.ValueOf(jo)
	log.Println("Got Tc_OBJECT %v\n", rv)
	if rv.Kind() == reflect.Ptr {
		name := rv.Elem().Type().Name()
		log.Println("jo type is %v", name)
	} else {
		log.Println("jo type is %v\n", rv.Type().Name())
	}
	return jo
}

func send() {
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second * 10)
		messageVO := vo.ReturnMessageVO{
			System:      "UYW",
			ProcessId:   strconv.Itoa(i),
			TaskId:      strconv.Itoa(i),
			DetailsId:   strconv.Itoa(i),
			ExeSeq:      "",
			Type:        "",
			Time:        time.Now().Format("2006-01-02 15:04:05"),
			Status:      "0",
			Drpcode:     "",
			Msg:         "",
			OutPutParam: "",
		}
		mq.SendMq(messageVO)
	}
}

func testFile() {
	// 这里换成实际的 SSH 连接的 用户名，密码，主机名或IP，SSH端口
	sshClient, err := util.Connect("afcat", "Afcat@0202", "10.24.24.11", 22)
	if err != nil {
		log.Fatal("连接创建失败", err)
	}
	defer sshClient.Close()
	sftpClient, err := util.GetSftpClient(sshClient)
	if err != nil {
		log.Fatal("创建sftp客户端失败")
	}
	defer sftpClient.Close()

	// 用来测试的本地文件路径 和 远程机器上的文件夹
	var localFilePath = "D:\\project\\go\\test.sh"
	var remoteDir = "/tmp/12345/"
	util.PushFile(sftpClient, localFilePath, remoteDir)
	// 用来测试的远程文件路径 和 本地文件夹
	var remoteFilePath = "/tmp/vmware-tools-distrib/lib/modules/source/legacy/pvscsi1.tar"
	var localDir = "D:\\project\\go"

	util.PullFile(sftpClient, remoteFilePath, localDir)
	//task, err := util.ExecTask(sshClient, fmt.Sprintf("/usr/bin/sh %s", "/tmp/12345/test.sh"))
	//if 0 == task {
	//	log.Println("执行成功")
	//} else {
	//	log.Println("远程执行脚本失败", err)
	//}
}
