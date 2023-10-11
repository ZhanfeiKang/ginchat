package models

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"gopkg.in/fatih/set.v0"
	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	FormId   int64  // 发送者
	TargetId int64  // 接收者
	Type     int    // 发送类型		1群聊，2私聊，3广播
	Media    int    // 消息格式     1文字，2表情包，3图片，4音频
	Content  string // 消息内容
	Pic      string
	Url      string
	Desc     string // 描述
	Amount   int    // 其他数字统计
}

func (table *Message) TableName() string {
	return "message"
}

type Node struct {
	Conn      *websocket.Conn
	DataQueue chan []byte
	GroupSets set.Interface
}

// 映射关系
var clientMap map[int64]*Node = make(map[int64]*Node, 0)

// 读写锁
var rwLocker sync.RWMutex

// 聊天需要：发送者ID，接收者ID，发送类型，发送的内容，消息类型
func Chat(writer http.ResponseWriter, request *http.Request) {
	// 1.获取参数 并 校验token等合法性

	// token := query.Get("token")
	query := request.URL.Query()
	id := query.Get("userId")
	userId, _ := strconv.ParseInt(id, 10, 64)
	// msgType := query.Get("type")
	// targetId := query.Get("targetId")
	// content := query.Get("content")
	isvalida := true // checkToken() 待补充。。。 需要将上面的token和userid传到数据库查一下是否ok
	conn, err := (&websocket.Upgrader{
		// token 校验
		CheckOrigin: func(r *http.Request) bool {
			return isvalida
		},
	}).Upgrade(writer, request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 2.获取连接
	node := &Node{
		Conn:      conn,
		DataQueue: make(chan []byte, 50),
		GroupSets: set.New(set.ThreadSafe),
	}

	// 3. 用户关系

	// 4. userID 跟 node 绑定 并且加锁
	rwLocker.Lock()
	clientMap[userId] = node
	rwLocker.Unlock()

	// 5.完成发送的逻辑
	go sendProc(node)
	// 6.完成接收的逻辑
	go recvProc(node)

	sendMsg(userId, []byte("欢迎进入聊天室~"))
}

func sendProc(node *Node) {
	for {
		select {
		case data := <-node.DataQueue:
			fmt.Println("[ws] sendMsg >>>>> msg: ", string(data))
			err := node.Conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

func recvProc(node *Node) {
	for {
		_, data, err := node.Conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		broadMsg(data)
		fmt.Println("[ws] <<<<< ", string(data))
	}
}

var udpsendChan chan []byte = make(chan []byte, 1024)

func broadMsg(data []byte) {
	udpsendChan <- data
}

func init() {
	go udpSendProc()
	go udpRecvProc()
	fmt.Println("init gorountine:")
}

// 完成udp数据发送协程
func udpSendProc() {
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(192, 168, 56, 1),
		Port: 3000,
	})
	defer conn.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		select {
		case data := <-udpsendChan:
			fmt.Println("udpSendProc..")
			_, err := conn.Write(data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

// 完成udp数据接收协程
func udpRecvProc() {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 3000,
	})
	defer conn.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		var buf [512]byte
		n, err := conn.Read(buf[0:])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("udpRecvProc data: ", string(buf[0:n]))
		dispatch(buf[0:n])
	}
}

// 后端调度逻辑处理
func dispatch(data []byte) {
	msg := Message{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	switch msg.Type {
	case 1: // 私信
		fmt.Println("dispatch data: ", string(data))
		sendMsg(msg.TargetId, data)
		// case 2: // 群发
		// 	sendGroupMsg()
		// case 3: // 广播
		// 	sendAllMsg()
		// case 4:
		// 	break
	}
}

func sendMsg(userId int64, msg []byte) {
	fmt.Println("sendMsg >>> userID: ", userId, "  msg: ", string(msg))
	rwLocker.Lock()
	node, ok := clientMap[userId]
	rwLocker.Unlock()
	if ok {
		node.DataQueue <- msg
	}
}
