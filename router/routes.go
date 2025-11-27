package router

import (
	"bell_best/controller"
	_ "bell_best/docs"
	"bell_best/logger"
	"bell_best/middlewares"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter() *gin.Engine {
	r := gin.New()
	r.Use(logger.GinLogger(), logger.GinRecovery(true))
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := r.Group("/api/v1")

	// 注册业务路由
	v1.POST("/signup", controller.SignUpHandler)
	// 登录
	v1.POST("/login", controller.LoginHandler)
	v1.GET("/posts/", controller.GetPostListHandler)
	// 根据帖子时间或分数获取帖子列表
	v1.GET("/posts2/", controller.GetPostListHandler2)

	v1.GET("/community", controller.CommunityHandler)
	v1.GET("/community/:id", controller.CommunityDetailHandler)
	v1.GET("/post/:id", controller.GetPostDetailHandler)

	v1.Use(middlewares.JWTAuthMiddleware()) // 认证JWT中间件

	{
		v1.POST("/post", controller.CreatePostHandler)
		// 投票
		v1.POST("/vote", controller.PostVoteController)
	}

	r.GET("/ping", middlewares.JWTAuthMiddleware(), func(c *gin.Context) {
		c.String(200, "请登录")
	})
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"message": "404 not found",
		})
	})
	return r
}
