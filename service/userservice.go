package service

import (
	"fmt"
	"ginchat/models"
	"ginchat/utils"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// GetUserList
// @Summary 所有用户
// @Tags 用户模块
// @Success 200 {string} json{"code","message"}
// @Router /user/getUserList [get]
func GetUserList(c *gin.Context) {
	data := make([]*models.UserBasic, 10)
	data = models.GetUserList()

	c.JSON(200, gin.H{
		"code":    0, // 0 成功    -1 失败
		"message": "用户列表获取成功！",
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
	user.Name = c.PostForm("name")
	password := c.PostForm("password")
	repassword := c.PostForm("repassword")

	if user.Name == "" || password == "" || repassword == "" {
		c.JSON(200, gin.H{
			"code":    -1, // 0 成功    -1 失败
			"message": "用户名或密码不能为空！",
			"data":    nil,
		})
		return
	}

	data := models.FindUserByName(user.Name)
	if data.Name != "" {
		c.JSON(200, gin.H{
			"code":    -1, // 0 成功    -1 失败
			"message": "该用户已被注册！",
			"data":    nil,
		})
		return
	}

	if password != repassword {
		c.JSON(200, gin.H{
			"code":    -1, // 0 成功    -1 失败
			"message": "两次密码不一致！",
			"data":    nil,
		})
		return
	}
	// user.PassWord = password
	salt := fmt.Sprintf("%06d", rand.Int31())
	user.Salt = salt
	user.PassWord = utils.MakePassword(password, salt) // 将用户输的的密码进行md5加密，存储到数据库中
	models.CreateUser(user)

	c.JSON(200, gin.H{
		"code":    0, // 0 成功    -1 失败
		"message": "新增用户成功！",
		"data":    user,
	})
}

// DeleteUser
// @Summary 删除用户
// @Tags 用户模块
// @param id query string false "id"
// @Success 200 {string} json{"code","message"}
// @Router /user/deleteUser [get]
func DeleteUser(c *gin.Context) {
	// 只做逻辑删除，不做物理删除
	user := models.UserBasic{}
	id, _ := strconv.Atoi(c.Query("id"))
	user.ID = uint(id)

	models.DeleteUser(user)
	c.JSON(200, gin.H{
		"code":    0, // 0 成功    -1 失败
		"message": "删除用户成功！",
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
		fmt.Println(err)
		c.JSON(200, gin.H{
			"code":    -1, // 0 成功    -1 失败
			"message": "修改参数不匹配",
			"data":    nil,
		})
	} else {
		models.UpdateUser(user)
		c.JSON(200, gin.H{
			"code":    0, // 0 成功    -1 失败
			"message": "修改用户成功！",
			"data":    nil,
		})
	}

}

// Login
// @Summary 用户登录
// @Tags 用户模块
// @param name query string false "用户名"
// @param password query string false "密码"
// @Success 200 {string} json{"code","message"}
// @Router /user/login [post]
func Login(c *gin.Context) {
	// name := c.Query("name")
	// password := c.Query("password")
	name := c.Request.FormValue("name")
	password := c.Request.FormValue("password")
	user := models.FindUserByName(name)
	if user.Name == "" {
		c.JSON(200, gin.H{
			"code":    -1, // 0 成功    -1 失败
			"message": "该用户不存在！",
			"data":    nil,
		})
		return
	}
	flag := utils.ValidPassword(password, user.Salt, user.PassWord)
	if !flag {
		c.JSON(200, gin.H{
			"code":    -1, // 0 成功    -1 失败
			"message": "密码不正确！",
			"data":    nil,
		})
		return
	}
	pwd := utils.MakePassword(password, user.Salt)
	data := models.FindUserByNameAndPwd(name, pwd)
	c.JSON(200, gin.H{
		"code":    0, // 0 成功    -1 失败
		"message": "登录成功",
		"data":    data,
	})
}

// 防止跨域站点的伪造请求
var upGrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func SendMsg(c *gin.Context) {
	ws, err := upGrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer func(ws *websocket.Conn) {
		err = ws.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(ws)

	MsgHandler(ws, c)
}

func MsgHandler(ws *websocket.Conn, c *gin.Context) {
	for {
		msg, err := utils.Subscribe(c, utils.PublishKey)
		if err != nil {
			fmt.Println(err)
			return
		}
		tm := time.Now().Format("2006-01-02 15:04:05")
		m := fmt.Sprintf("[ws][%s]:%s", tm, msg)
		err = ws.WriteMessage(1, []byte(m))
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("消息发送成功！")
	}
}

func SendUserMsg(c *gin.Context) {
	models.Chat(c.Writer, c.Request)
}

func SearchFriends(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))
	users := models.SearchFriend(uint(userId))

	// c.JSON(200, gin.H{
	// 	"code":    0, // 0 成功    -1 失败
	// 	"message": "查询好友列表成功！",
	// 	"data":    users,
	// })
	utils.RespOKList(c.Writer, users, len(users))
}
