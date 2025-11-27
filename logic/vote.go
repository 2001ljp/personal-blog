package logic

import (
	"bell_best/dao/redis"
	"bell_best/models"
	"go.uber.org/zap"

	"strconv"
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

// VoteForPost 为帖子投票的函数
func VoteForPost(userID int64, p *models.ParamVoteData) (err error) {
	zap.L().Debug("VoteForPost", zap.Int64("user_id", userID), zap.String("post_id", p.PostID), zap.Int8("direction", p.Direction))
	return redis.VoteForPost(strconv.Itoa(int(userID)), p.PostID, float64(p.Direction))
}
