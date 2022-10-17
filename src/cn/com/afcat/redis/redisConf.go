package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"log"
	"time"
)

var sf *redis.FailoverOptions

var ctx context.Context = context.Background()
var rdb *redis.Client

func RedisConf(sentinelAddrs []string, password string) {
	sf = &redis.FailoverOptions{
		// The master name.
		MasterName: "mymaster",
		// A seed list of host:port addresses of sentinel nodes.
		SentinelAddrs: sentinelAddrs,

		// Following options are copied from Options struct.
		Password: password,
		DB:       0,
	}
}

func GetData(key string) []byte {
	value := GetRedisClient().Get(ctx, key)
	if value.Err() != nil {
		log.Printf("查询redis数据失败", value.Err())
		panic("查询redis数据失败")
	}
	bytes, err := value.Bytes()
	if err != nil {
		log.Printf("查询redis数据失败", err)
		panic("查询redis数据失败")
	}
	return bytes
}

func SetData(key string, value []byte) {
	set := GetRedisClient().Set(ctx, key, value, 0)
	if set.Err() != nil {
		log.Printf("保存redis数据失败", set.Err())
		panic("保存redis数据失败" + key)
	}
}

func SetExData(key string, value []byte, expiration time.Duration) {
	set := GetRedisClient().Set(ctx, key, value, expiration)
	if set.Err() != nil {
		log.Printf("保存redis数据失败", set.Err())
		panic("保存redis数据失败" + key)
	}
}

func Exists(key string) bool {
	exists, err := GetRedisClient().Exists(ctx, key).Result()
	if err != nil {
		log.Printf("查询redis数据是否存在失败", err)
		panic(err)
	}
	if exists > 0 {
		return true
	} else {
		return false
	}
}

func Del(key string) {
	_, err := GetRedisClient().Del(ctx, key).Result()
	if err != nil {
		log.Printf("删除redis数据失败", err)
		panic(err)
	}
}

func GetRedisClient() *redis.Client {
	if rdb != nil {
		return rdb
	}
	defer func() {
		if err := recover(); err != nil {
			log.Println("redis获取连接失败,3秒后重连，%v")
			if rdb != nil {
				rdb.Close()
			}
			rdb = nil
			GetRedisClient()
		}
	}()
	rdb = redis.NewFailoverClient(sf)
	return rdb
}

func Keys(key string) []string {
	keys := GetRedisClient().Keys(ctx, key)
	if keys.Err() != nil {
		log.Printf("保存redis数据失败", keys.Err())
	}
	return keys.Val()
}
