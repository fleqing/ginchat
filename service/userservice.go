package service

import (
	"fmt"
	"ginchat/models"
	"ginchat/utils"
	"html/template"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// GetUserList
// @Summary 获取用户列表
// @Tags 用户模块
// @Success 200 {string} json{"code","message"}
// @Router /user/getUserList [get]
func GetUserList(c *gin.Context) {
	// data := make([]*models.UserBasic, 10)
	data := models.GetUserList()
	c.JSON(200, gin.H{
		"code":    "0", //0表示成功,-1表示失败
		"message": "获取成功",
		"data":    data,
	})
}

// CreateUser
// @Summary 新增用户
// @Tags 用户模块
// @param name query string false "用户名"
// @param password query string false "密码"
// @param repassword query string false "确认密码"
// @Success 200 {string} json{"code","message"}
// @Router /user/createUser [get]
func CreateUser(c *gin.Context) {
	user := models.UserBasic{}
	user.Name = c.Request.FormValue("name")
	password := c.Request.FormValue("password")
	repassword := c.Request.FormValue("Identity")

	salt := fmt.Sprintf("%06d", rand.Intn(1000))

	// user.LoginTime = time.Now()
	// user.LoginOutTime = time.Now()
	// user.HearBeatTime = time.Now()

	data := models.FindUserByName(user.Name)
	if user.Name == "" || password == "" || repassword == "" {
		c.JSON(200, gin.H{
			"code":    "-1", //0表示成功,-1表示失败
			"message": "用户名或密码不能为空",
			"data":    user,
		})
		return
	}

	if data.Name != "" {
		c.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失败
			"message": "用户名已注册！",
			"data":    user,
		})
		return
	}

	if password != repassword {
		c.JSON(200, gin.H{
			"code":    "-1", //0表示成功,-1表示失败
			"message": "两次密码不一致",
			"data":    user,
		})
		return
	}
	// user.PassWord = password
	user.PassWord = utils.MakePassword(password, salt)
	user.Salt = salt
	models.CreateUser(user)
	c.JSON(200, gin.H{
		"code":    "0", //0表示成功,-1表示失败
		"message": "新增成功",
		"data":    user,
	})
}

// CreateUser
// @Summary 删除用户
// @Tags 用户模块
// @param id query string false "id"
// @Success 200 {string} json{"code","message"}
// @Router /user/deleteUser [get]
func DeleteUser(c *gin.Context) {
	user := models.UserBasic{}
	id, _ := strconv.Atoi(c.Query("id"))
	user.ID = uint(id)
	models.DeleteUser(user)
	c.JSON(200, gin.H{
		"code":    "0", //0表示成功,-1表示失败
		"message": "删除成功",
		"data":    user,
	})
}

// UpdateUser
// @Summary 修改用户
// @Tags 用户模块
// @param id formData string false "id"
// @param name formData string false "name"
// @param password formData string false "password"
// @param phone formData string false "phone"
// @param email formData string false "email"
// @Success 200 {string} json{"code","message"}
// @Router /user/updateUser [post]
func UpdateUser(c *gin.Context) {
	user := models.UserBasic{}
	id, _ := strconv.Atoi(c.PostForm("id"))
	user.ID = uint(id)
	user.Name = c.PostForm("name")
	user.PassWord = c.PostForm("password")
	user.Phone = c.PostForm("phone")
	user.Email = c.PostForm("email")
	_, err := govalidator.ValidateStruct(user)
	if err != nil {
		c.JSON(200, gin.H{
			"code":    "-1", //0表示成功,-1表示失败
			"message": "参数错误",
			"data":    user,
		})
	} else {
		models.UpdateUser(user)
		c.JSON(200, gin.H{
			"code":    "0", //0表示成功,-1表示失败
			"message": "修改成功",
			"data":    user,
		})
	}
}

// CreateUser
// @Summary 用户登录
// @Tags 用户模块
// @param name query string false "用户名"
// @param password query string false "密码"
// @Success 200 {string} json{"code","message"}
// @Router /user/findUserByNameAndPassWord [post]
func FindUserByNameAndPassWord(c *gin.Context) {
	data := models.UserBasic{}
	// name := c.Query("name")
	// password := c.Query("password")
	name := c.Request.FormValue("name")
	password := c.Request.FormValue("password")

	user := models.FindUserByName(name)
	if user.Name == "" {
		c.JSON(200, gin.H{
			"code":    -1, //0表示成功,-1表示失败
			"message": "用户不存在",
			"data":    data,
		})
		return
	}

	flag := utils.ValidPassword(password, user.Salt, user.PassWord)

	if !flag {
		c.JSON(200, gin.H{
			"code":    -1, //0表示成功,-1表示失败
			"message": "密码错误",
			"data":    data,
		})
		return
	}

	data = models.FindUserByNameAndPassWord(name, utils.MakePassword(password, user.Salt))
	c.JSON(200, gin.H{
		"code":    0, //0表示成功,-1表示失败
		"message": "登录成功",
		"data":    data,
	})
}

// 防止跨域站点伪造请求
// 在你的代码中，你创建了一个 websocket.Upgrader 的实例 upGrade，并设置了 CheckOrigin 字段。
// CheckOrigin 是一个函数，它接受一个 *http.Request 参数，返回一个布尔值，用来检查源（即发起请求的域）是否被允许。
// 如果 CheckOrigin 返回 true，则请求被接受，如果返回 false，则请求被拒绝。
var upGrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func SendMsg(c *gin.Context) {
	// 将 HTTP 连接升级为 WebSocket 连接
	ws, err := upGrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("err", err)
		return
	}
	// (ws) 是对匿名函数的参数进行调用。否则，匿名函数将不会被调用。
	defer func(ws *websocket.Conn) {
		err := ws.Close()
		if err != nil {
			fmt.Println("err", err)
		}

	}(ws)
	MsgHandler(ws, c)
}

func MsgHandler(ws *websocket.Conn, c *gin.Context) {
	msg, err := utils.Subscribe(c, utils.PublishKey)
	if err != nil {
		fmt.Println("err", err)
	}
	tm := time.Now().Format("2006-01-02 15:04:05")
	m := fmt.Sprintf("[ws][%s] %s", tm, msg)
	err = ws.WriteMessage(1, []byte(m))
	if err != nil {
		fmt.Println("err", err)
	}
}

func SendUserMsg(c *gin.Context) {
	models.Chat(c.Writer, c.Request)
}

func ToRegister(c *gin.Context) {
	ind, err := template.ParseFiles("views/user/register.html")
	if err != nil {
		panic(err)
	}
	ind.Execute(c.Writer, "register")
}

func SearchFriends(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))
	users := models.SearchFrend(uint(userId))
	// c.JSON(200, gin.H{
	// 	"code":    0,
	// 	"message": "获取成功",
	// 	"data":    users,
	// })
	utils.RespOKList(c.Writer, users, len(users))
}

func AddFriend(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))
	targetName := c.Request.FormValue("targetName")
	flag, msg := models.AddFriend(uint(userId), targetName)
	if flag == 0 {
		utils.RespOk(c.Writer, flag, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}

func CreateCommunity(c *gin.Context) {
	community := models.Community{}
	ownerId, _ := strconv.Atoi(c.Request.FormValue("ownerId"))
	community.Name = c.Request.FormValue("name")
	community.OwnerId = uint(ownerId)
	flag, msg := models.CreateCommunity(community)
	if flag == 0 {
		utils.RespOk(c.Writer, flag, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}

func LoadCommunity(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("ownerId"))
	community := models.LoadCommunity(uint(userId))
	utils.RespOKList(c.Writer, community, len(community))
}

func JoinGroup(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))
	communityId := c.Request.FormValue("comId")
	flag, msg := models.JoinGroup(uint(userId), communityId)
	if flag == 0 {
		utils.RespOk(c.Writer, flag, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}

func RedisMsg(c *gin.Context) {
	userIdA, _ := strconv.Atoi(c.Request.FormValue("userIdA"))
	userIdB, _ := strconv.Atoi(c.Request.FormValue("userIdB"))
	start, _ := strconv.ParseInt(c.Request.FormValue("start"), 10, 64)
	end, _ := strconv.ParseInt(c.Request.FormValue("end"), 10, 64)
	isRev, _ := strconv.ParseBool(c.Request.FormValue("isRev"))
	msgs := models.RedisMsg(int64(userIdA), int64(userIdB), start, end, isRev)
	utils.RespOKList(c.Writer, "ok", msgs)
}
