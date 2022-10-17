package task

import (
	"bytes"
	"fmt"
	"go-agent/src/cn/com/afcat/prop"
	"go-agent/src/cn/com/afcat/redis"
	"go-agent/src/cn/com/afcat/util"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var wg sync.WaitGroup

func HeartCheack() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("主机心跳检测失败，%v", err)
			time.Sleep(time.Second)
			HeartCheack()
		}
	}()
	date := time.Now()
	for {
		hostHeart := redis.GetData("heartInterval")
		var heartInterval int = 10000
		if hostHeart != nil && string(hostHeart) != "" {
			atoi, err := strconv.Atoi(string(hostHeart))
			if err != nil {
				log.Println("主机心跳检查间隔时间查询失败")
			}
			heartInterval = atoi
		}
		sub := time.Now().Sub(date)
		second := int(sub.Seconds())
		if second > (heartInterval / 1000 / 3) {
			keys := redis.Keys("heart:host:*")
			if keys != nil && len(keys) > 0 {
				//计算需要开启多少个检测任务，每个检测任务检测个数
				count := 10
				end := (len(keys) / count) + 1
				log.Println("将开启线程数:%v", end)
				for i := 0; i < end; i++ {
					date = time.Now()
					wg.Add(1)
					go HostPingTask(keys, i, count, heartInterval)
				}
				wg.Wait()
			}
		} else {
			time.Sleep(time.Second)
		}
	}

}

func HostPingTask(host []string, start int, count int, heartInterval int) {
	defer func() {
		log.Println("执行心跳检查结束：%d %d", start, count)
		wg.Done()
	}()
	//goos := runtime.GOOS
	end := (start + 1) * count
	if end > len(host) {
		end = len(host)
	}
	fire := start * count
	log.Println("执行：%d %d", start, count)

	prop.ManageIps.Traverse()
	for i := fire; i < end; i++ {
		if len(host) > i {
			defer func() {
				if err := recover(); err != nil {
					log.Println("主机心跳检测失败，%v", err)
				}
			}()
			hostKey := host[i]
			//获取检测IP的状态
			hostStatusVO := GetData(hostKey)
			if hostStatusVO != nil {
				jsonMap := hostStatusVO.JsonMap()
				//获取要检测对象上次的代理机
				hostStatusMap := jsonMap.(map[string]interface{})
				//获取要检测对象上次的代理机
				agent := hostStatusMap["agent"].(string)
				//获取上次的心跳时间
				lastTimeString := hostStatusMap["time"].(string)
				var lastTime time.Time
				if &lastTimeString != nil && "" != lastTimeString {
					location, err := time.ParseInLocation("2006-01-02 15:04:05", lastTimeString, time.Local)
					if err != nil {
						log.Println("时间转换失败，%s", err)
					}
					lastTime = location
				}
				ip := hostStatusMap["ip"].(string)
				if &ip == nil {
					continue
				}

				if strings.EqualFold(prop.Ip, ip) {
					saveToRedis(hostStatusVO, hostKey)
					continue
				}
				if strings.Index(ip, ".") != -1 && prop.ManageIps.Size() > 0 {
					//rs := []rune(ip)
					//string(rs[0:strings.LastIndex(ip, ".")])
					ipp := ip[0:strings.LastIndex(ip, ".")]
					flag := prop.ManageIps.Exist(ipp)
					if !flag {
						ipp = ipp[0:strings.LastIndex(ip, ".")]
						flag = prop.ManageIps.Exist(ipp)
					}
					if flag {
						log.Println("匹配到网段%s", ipp)
						mangeIp := redis.GetData(ip + ":Hear:ip")
						//判断要检测IP的管理网段高于本网段
						if mangeIp == nil || strings.Contains(string(mangeIp), ipp) {
							redis.SetData(ip+":Hear:ip", []byte(ipp))
						} else {
							continue
						}
						hostCheck(agent, lastTime, heartInterval, ip, hostStatusVO, hostKey)
					}
				} else {
					hostCheck(agent, lastTime, heartInterval, ip, hostStatusVO, hostKey)
				}
			}
		}
	}
}

func hostCheck(agent string, lastTime time.Time, heartInterval int, ip string, hostStatusVO *util.JavaTcObject, hostKey string) {
	systemTime, err := time.ParseInLocation("2006-01-02 15:04:05", string(redis.GetData("systemTime")), time.Local)
	if err != nil {
		log.Println("系统时间转换失败，%s", err)
		systemTime = time.Now()
	}
	if &agent == nil || strings.EqualFold(agent, "") || int(systemTime.Sub(lastTime)) > heartInterval {
		ping := Ping(ip)
		if ping == 0 {
			saveToRedis(hostStatusVO, hostKey)
		}
	} else if strings.EqualFold(agent, prop.AgentId) {
		saveToRedis(hostStatusVO, hostKey)
	}
}

func saveToRedis(hostStatusVO *util.JavaTcObject, key string) {
	//设置本代理的版本
	classes := hostStatusVO.Classes
	fields := classes[0].Fields
	for _, v := range fields {
		if strings.EqualFold(v.FieldName, "time") {
			v.FieldValue = string(redis.GetData("systemTime"))
		}
		if strings.EqualFold(v.FieldName, "version") {
			v.FieldValue = prop.AgentVersion
		}
		if strings.EqualFold(v.FieldName, "agent") {
			v.FieldValue = prop.AgentId
		}
	}
	buffer := new(bytes.Buffer)
	util.SerializeJavaEntity(buffer, hostStatusVO)
	redis.SetData(key, buffer.Bytes())
}

func GetData(key string) *util.JavaTcObject {
	bytes1 := redis.GetData(key)

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
	/*rv := reflect.ValueOf(jo)
	if rv.Kind() == reflect.Ptr {
		name := rv.Elem().Type().Name()
		log.Println("jo type is %v", name)
	} else {
		log.Println("jo type is %v\n", rv.Type().Name())
	}*/
	return jo
}

func Ping(dstIP string) int {
	//log.Println("ping %s", dstIP)
	//关键代码
	cmd := exec.Command("ping", "-c", "3", dstIP)
	//cmd := exec.Command(fmt.Sprintf("timeout 1 bash -c 'cat < /dev/null > /dev/tcp/%s/%s'", dstIP, util.Port))
	err := cmd.Run()
	if err != nil {
		log.Println("ping %s failed. err=", dstIP, err)
		return 1
	}
	return 0
}

func Tcp(dstIP string) int {
	//cmd := exec.Command("ping", "-c", count, dstIP)
	cmd := exec.Command(fmt.Sprintf("timeout 1 bash -c 'cat < /dev/null > /dev/tcp/%s/%s'", dstIP, prop.Port))
	err := cmd.Run()
	if err != nil {
		//log.Println("ping %s failed. err=", dstIP, err)
		return 1
	}
	return 0
}
