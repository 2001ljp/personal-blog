package redis

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"math"
	"strconv"
	"time"
)

// 基于用户投票的相关算法:http://www.ruanyifeng.com/blog/algorithm/
// 本项目使用简化版的投票分数
// 投一票就加432分	86400/200 -> 200张赞成票可以给你的帖子续一天

/* 投票的几种情况
direction=1时，有两种情况：
	1. 之前没有投过票，现在投赞成票--> 更新分数和投票记录	差值绝对值是1	+432
	2. 之前投反对票，现在改赞成票--> 更新分数和投票记录	差值绝对值是2	+432*2
direction=0时，有两种情况：
	1. 之前投反对票，现在取消投票--> 更新分数和投票记录	差值绝对值是1	+432
	2. 之前投赞成票，现在取消投票--> 更新分数和投票记录	差值绝对值是1	-432
direction=-1时，有两种情况：
	1. 之前没有投过票，现在投反对票--> 更新分数和投票记录	差值绝对值是1	-432
	2. 之前投赞成票，现在改反对票--> 更新分数和投票记录		差值绝对值是2	-432*2

投票的限制：
每个帖子自发帖之日起只在一个星期内允许投票
	1.到期之后将redis中保存的赞成票数及反对票数存储到mysql表中
	2.到期之后删除那个 KeyPostVotedPF
*/

const (
	oneWeekInSeconds = 7 * 24 * 60 * 60
	scorePerVote     = 432 // 每一票值多少分
)

var (
	ErrVoteTimeExpired = errors.New("vote time expired")
	ctx                = context.Background()
	ErrVoteRepested    = errors.New("vote repested")
)

func CreatePost(postID, communityID int64) error {
	pipeline := client.TxPipeline()
	// 帖子时间
	client.ZAdd(ctx, GetRedisKey(KeyPostTime), &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: postID,
	})
	// 帖子分数
	client.ZAdd(ctx, GetRedisKey(KeyPostScore), &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: postID,
	})
	// 把帖子id加到社区的set
	cKey := GetRedisKey(KeyCommunityPF + strconv.Itoa(int(communityID)))
	pipeline.SAdd(ctx, cKey, postID)
	_, err := pipeline.Exec(ctx)
	return err
}

// VoteForPost 为帖子投票的函数
func VoteForPost(userID, postID string, value float64) error {
	// 1. 判断投票限制
	// 去redis取帖子发布时间
	ctx := context.Background()
	postTime := client.ZScore(ctx, GetRedisKey(KeyPostTime), postID).Val()
	if float64(time.Now().Unix())-postTime > oneWeekInSeconds {
		return ErrVoteTimeExpired
	}
	// 2和3需要放在一个pipeline事务中操作

	// 2. 更新帖子分数
	// 先查当前用户给当前帖子的投票记录
	ov := client.ZScore(ctx, GetRedisKey(KeyPostVotedPF+postID), userID).Val()
	// 如果这一次投票的值和之前保存的值一致，就提示不允许重复投票
	if value == ov {
		return ErrVoteRepested
	}
	var op float64
	if value > ov {
		op = 1
	} else {
		op = -1
	}
	diff := math.Abs(ov - value)
	pipeline := client.TxPipeline()                                                         // 计算两次投票的差值
	pipeline.ZIncrBy(ctx, GetRedisKey(KeyPostScore), op*diff*scorePerVote, postID).Result() // 更新分数
	// 3. 记录用户为该帖子投票的数据
	if value == 0 {
		pipeline.ZRem(ctx, GetRedisKey(KeyPostVotedPF+postID), userID)
	} else {
		pipeline.ZAdd(ctx, GetRedisKey(KeyPostVotedPF+postID), &redis.Z{
			Score:  value, // 赞成票还是反对票
			Member: userID,
		})
	}
	_, err := pipeline.Exec(ctx)
	return err
}
