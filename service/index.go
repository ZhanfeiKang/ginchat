package service

import (
	"ginchat/models"
	"html/template"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetIndex
// @Tags 首页
// @Success 200 {string} welcome
// @Router /index [get]
func GetIndex(c *gin.Context) {
	ind, err := template.ParseFiles("index.html", "views/chat/head.html")
	if err != nil {
		panic(err)
	}
	ind.Execute(c.Writer, "index")
	// if err != nil {
	// 	panic(err)
	// }
}

func ToRegister(c *gin.Context) {
	ind, err := template.ParseFiles("views/user/register.html")
	if err != nil {
		panic(err)
	}
	ind.Execute(c.Writer, nil)
}

func ToChat(c *gin.Context) {
	ind, err := template.ParseFiles("views/chat/index.html",
		"views/chat/head.html",
		"views/chat/foot.html",
		"views/chat/tabmenu.html",
		"views/chat/concat.html",
		"views/chat/group.html",
		"views/chat/profile.html",
		"views/chat/main.html",
		"views/chat/createcom.html",
		"views/chat/userinfo.html")
	if err != nil {
		panic(err)
	}

	userId, _ := strconv.Atoi(c.Query("userId"))
	token := c.Query("token")

	user := models.UserBasic{}
	user.ID = uint(userId)
	user.Identity = token

	ind.Execute(c.Writer, user)
}

func Chat(c *gin.Context) {
	models.Chat(c.Writer, c.Request)
}
