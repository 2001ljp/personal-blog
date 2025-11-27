package redis

import (
	"bell_best/models"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
)

func getIDsFormKey(key string, page, size int64) ([]string, error) {
	start := (page - 1) * size
	end := start + size - 1
	// 3. ZRevRange 按分数从大到小的顺序查询指定数量的id
	return client.ZRevRange(ctx, key, start, end).Result()
}

func GetPostIDsInOrder(p *models.ParamPostList) ([]string, error) {
	// 从redis获取id
	// 1. 根据用户请求中携带的order参数确定要查询的redis key
	key := GetRedisKey(KeyPostTime)
	if p.Order == models.OrderScore {
		key = GetRedisKey(KeyPostScore)
	}
	return getIDsFormKey(key, p.Page, p.Size)
}

// GetPostVoteData 根据ids查询每篇帖子的投赞成票的数据
func GetPostVoteData(ids []string) (data []int64, err error) {
	//data = make([]int64, 0,len(ids))
	//for _, id := range ids {
	//	key := GetRedisKey(KeyPostVotedPF + id)
	//	v1 := client.ZCount(ctx, key, "1", "1").Val()
	//	data = append(data, v1)

	// 使用pipeline一次发送多条命令减少RTT
	pipeline := client.Pipeline()
	for _, id := range ids {
		key := GetRedisKey(KeyPostVotedPF + id)
		pipeline.ZCount(ctx, key, "1", "1")
	}
	cmders, err := pipeline.Exec(ctx)
	if err != nil {
		return nil, err
	}
	data = make([]int64, 0, len(ids))
	for _, cmder := range cmders {
		v := cmder.(*redis.IntCmd).Val()
		data = append(data, v)
	}
	return
}

// GetCommunityPostIDsInOrder 按社区查询ids
func GetCommunityPostIDsInOrder(p *models.ParamPostList) ([]string, error) {
	orderKey := GetRedisKey(KeyPostTime)
	if p.Order == models.OrderScore {
		orderKey = GetRedisKey(KeyPostScore)
	}
	// 使用zinterstore 把分区的帖子set与帖子分数的zset 生成一个新的zset
	// 针对新的zset按之前的逻辑取数据

	// 社区的key
	cKey := GetRedisKey(KeyCommunityPF + strconv.Itoa(int(p.CommunityID)))
	// 利用缓存key减少zinterstore执行的次数
	key := orderKey + strconv.Itoa(int(p.CommunityID))
	if client.Exists(ctx, key).Val() < 1 {
		// 不存在，需要计算
		pipeline := client.Pipeline()
		pipeline.ZInterStore(ctx, key, &redis.ZStore{
			Aggregate: "MAX",
			Keys:      []string{cKey, orderKey},
		}) // zinterstore 计算
		pipeline.Expire(ctx, key, 60*time.Second) // 设置超时时间
		_, err := pipeline.Exec(ctx)
		if err != nil {
			return nil, err
		}
	}
	// 存在的话就直接根据key查询ids
	return getIDsFormKey(key, p.Page, p.Size)
}
