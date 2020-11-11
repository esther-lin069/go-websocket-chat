package main

import (
	"github.com/go-redis/redis"
)

/*獲取Redis連線*/
func GetRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
}

/*獲取Redis連線*/
func GetRedisForPrivate() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       1,
	})
}

/*Redis存取+刪除*/
func (c *Client) ZsetMessage(m RedisMsg) {
	rdb := c.redis_conn

	msg := redis.Z{
		Score:  m.Id,
		Member: m.Value,
	}

	length := rdb.ZCard(m.User).Val()
	if length >= zrange {
		//data := c.zrangeMessage(m.User, zrange/2)
		//將前一半筆放入mysql
		//PutMsgList(m.User, data)
		rdb.ZRemRangeByRank(m.User, 0, zrange/2)
	}

	err := rdb.ZAdd(m.User, msg).Err()
	if err != nil {
		panic(err)
	}
}

func (c *Client) ZrangeMessage(id string, len int64) []redis.Z {
	rdb := c.redis_conn

	data, err := rdb.ZRangeWithScores(id, 0, len).Result()
	if err != nil {
		panic(err)
	}
	return data

}

func DelKey(key string) {
	rdb := GetRedisClient()
	err := rdb.Del(key).Err()
	if err != nil {
		panic(err)
	}
}

func HsetForPrivate(key, field, val string) {
	rdb := GetRedisForPrivate()
	defer rdb.Close()

	err := rdb.HSet(key, field, val)
	if err != nil {
		return
	}
}

func GetHashForPrivate(key string) map[string]string {
	rdb := GetRedisForPrivate()
	defer rdb.Close()

	result, _ := rdb.HGetAll(key).Result()

	return result

}
