package router

import (
	"github.com/gin-gonic/gin"

	ginSwagger "github.com/swaggo/gin-swagger"

	swaggerfiles "github.com/swaggo/files"

	docs "ginchat/docs"
	"ginchat/service"
)

func Router() *gin.Engine {
	r := gin.Default()
	docs.SwaggerInfo.BasePath = ""
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// 静态资源
	// 这行代码将 URL 路径 /asset 映射到本地的 asset/ 目录。
	// 这意味着，如果你在浏览器中访问 http://yourserver.com/asset/somefile.jpg，Gin .
	//会在 asset/ 目录下查找 somefile.jpg 文件并返回
	r.Static("/asset", "asset/")
	/*
		这行代码告诉 Gin 在 views/ 目录及其所有子目录下加载所有的 HTML 文件。这些文件可以被用作模板，用于生成动态的 HTML 页面。
		是一个 glob 模式，表示匹配任何文件和子目录。
	*/
	r.LoadHTMLGlob("views/**/*")
	// 首页
	r.GET("/toRegister", service.ToRegister)
	r.GET("/", service.GetIndex)
	r.GET("/index", service.GetIndex)
	r.GET("/toChat", service.ToChat)
	r.GET("/chat", service.Chat)
	// 加载好友列表
	r.POST("/searchFriends", service.SearchFriends)
	//用户模块
	r.POST("/user/getUserList", service.GetUserList)
	r.POST("/user/createUser", service.CreateUser)
	r.POST("/user/deleteUser", service.DeleteUser)
	r.POST("/user/updateUser", service.UpdateUser)
	r.POST("/user/findUserByNameAndPwd", service.FindUserByNameAndPassWord)
	// 发送消息
	r.GET("/user/sendMsg", service.SendMsg)
	// 发送用户消息
	r.GET("/user/sendUserMsg", service.SendUserMsg)
	// 上传文件
	r.POST("/attach/upload", service.Upload)
	// 添加好友
	r.POST("/contact/addfriend", service.AddFriend)
	// 创建群聊
	r.POST("/contact/createCommunity", service.CreateCommunity)
	// 加载群列表
	r.POST("contact/loadcommunity", service.LoadCommunity)
	// 添加群
	r.POST("contact/joinGroup", service.JoinGroup)
	// 加载消息
	r.POST("/user/redisMsg", service.RedisMsg)
	return r
}
