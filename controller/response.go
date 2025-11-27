package controller

import "github.com/gin-gonic/gin"

// response：响应
// 封装响应

/*
{
	"code": 10001, // 程序中的错误码
	"msg": xx, // 提示信息
	"data": {}, // 数据
}
*/

type ResponseDate struct {
	Code ResCode     `json:"code"`
	Msg  interface{} `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func ResponseError(c *gin.Context, code ResCode) {
	c.JSON(200, &ResponseDate{
		Code: code,
		Msg:  code.Msg(),
		Data: nil,
	})
}
func ResponseErrorWithMsg(c *gin.Context, code ResCode, msg interface{}) {
	c.JSON(200, &ResponseDate{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}

func ResponseSuccess(c *gin.Context, data interface{}) {
	c.JSON(200, &ResponseDate{
		Code: CodeSuccess,
		Msg:  CodeSuccess.Msg(),
		Data: data,
	})
}
