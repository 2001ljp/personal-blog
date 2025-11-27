package logic

import (
	"bell_best/dao/mysql"
	"bell_best/models"
	"bell_best/pkg/jwt"
	"bell_best/pkg/snowflake"
)

// 存放业务逻辑的代码

func SignUp(p *models.ParamSignUp) (err error) {
	// 判断注册用户存不存在
	if err := mysql.CheckUserExist(p.Username); err != nil {
		return err
	}
	// 生成UID
	userID := snowflake.GenID()
	// 构造一个User实例
	user := &models.User{
		UserID:   userID,
		Username: p.Username,
		Password: p.Password,
	}
	// 保存进数据库
	return mysql.InsertUser(user)
}

func Login(p *models.ParamLogin) (user *models.User, err error) {
	user = &models.User{
		Username: p.Username,
		Password: p.Password,
	}
	// 传递的是指针，就能拿到userID
	if err := mysql.Login(user); err != nil {
		return nil, err
	}
	// 生产JWT
	token, err := jwt.GenToken(user.UserID, user.Username)
	user.Token = token
	return
}
