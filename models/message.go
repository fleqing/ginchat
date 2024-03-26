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
	FromId   int64  //发送者
	TargetId int64  //接收者
	Type     int    //发送类型 群聊、私聊、广播
	Media    int    //消息类型 文本、图片、视频、文件
	Content  string //消息内容
	Pic      string
	Url      string
	Desc     string
	Amount   int //其他数字统计
}

func (table *Message) TableName() string {
	return "message"
}

type Node struct {
	Conn      *websocket.Conn
	DataQueue chan []byte
	//GroupSets 可能被用来存储用户加入的聊天组的 ID
	GroupSets set.Interface
}

// 映射关系
var clientMap map[int64]*Node = make(map[int64]*Node, 0)

// 读写锁
var rwlocker sync.RWMutex

// 处理HTTP请求，建立WebSocket连接，并将用户的连接信息存储在 clientMap 中，
// 然后启动两个协程进行消息的发送和接收。
func Chat(writer http.ResponseWriter, request *http.Request) {
	//检验token
	query := request.URL.Query()
	// token := query.Get("token")
	Id := query.Get("userId")
	userId, _ := strconv.ParseInt(Id, 10, 64)
	// msgType := query.Get("type")
	// targetId := query.Get("targetId")
	// contex := query.Get("content")
	isvalida := true
	conn, err := (&websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return isvalida
		},
	}).Upgrade(writer, request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	// 创建一个 Node 实例,存储websocket连接
	node := &Node{Conn: conn,
		DataQueue: make(chan []byte, 50),
		GroupSets: set.New(set.ThreadSafe)}
	// 将用户的连接信息存储到map中
	rwlocker.Lock()
	clientMap[userId] = node
	rwlocker.Unlock()
	// 开启协程，进行消息的发送
	go sendProc(node)
	// 开启协程，进行消息的接收
	go recvProc(node)
	sendMsg(userId, []byte("欢迎进入聊天系统"))
}

// 从 Node 的 DataQueue 中读取数据，并将这些数据通过 WebSocket 连接发送出去
// 从 node.DataQueue 中读取数据，并将数据作为WebSocket消息发送出去
func sendProc(node *Node) {
	for {
		select {
		case data := <-node.DataQueue:
			fmt.Println("[ws] sendProc >>>> msg:", string(data))
			err := node.Conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
	// for data := range node.DataQueue {
	// 	fmt.Println("[ws] sendProc >>>> msg:", string(data))
	// 	err := node.Conn.WriteMessage(websocket.TextMessage, data)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		return
	// 	}
	// }
}

// 从WebSocket连接中读取消息，并将消息数据发送到广播通道
func recvProc(node *Node) {
	for {
		_, data, err := node.Conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		// 将消息放在广播通道
		broadMsg(data)
		fmt.Println("[ws] recvproc <<<<<<<<<<", string(data))
	}
}

var udpsendChan chan []byte = make(chan []byte, 1024)

func broadMsg(data []byte) {
	udpsendChan <- data
	// fmt.Println("broad_time", time.Now())
	// print("broadMsg >>>", string(data))
}

// 自动调用
func init() {
	go udpSendProc()
	go udpRecvProc()
	fmt.Println("init goroutine")
}

// 发送协程
func udpSendProc() {
	// 创建一个 UDP 连接,连接的目标地址是 192.168.0.255
	con, err := net.DialUDP("udp", nil, &net.UDPAddr{
		// TODO IP地址可能需要修改
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 3001,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer con.Close()
	// fmt.Println("udp_time", time.Now())
	for {
		select {
		case data := <-udpsendChan:
			fmt.Println("udpSendProc      ", string(data))
			_, err := con.Write(data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
	// for data := range udpsendChan {
	// 	data = <-udpsendChan // Replace select statement with channel receive operation
	// 	fmt.Println("udpSendProc      ", string(data))
	// 	// 将从通道读取的数据发送到 UDP 连接。
	// 	_, err := con.Write(data)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		return
	// 	}
	// }
}

// 接收协程
func udpRecvProc() {
	con, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 3001,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer con.Close()
	for {
		var buf [1024]byte
		n, err := con.Read(buf[:])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("udpRecvProc      ", string(buf[:n]))
		dispatch(buf[:n])
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
	// 私聊
	case 1:
		fmt.Println("dispatch", string(data))
		sendMsg(msg.TargetId, data)
	}
}

// 将数据放在对应的用户的 DataQueue 中
func sendMsg(userId int64, data []byte) {
	fmt.Println("sendmsg >>> userid", userId, "msg", string(data))
	rwlocker.RLock()
	node, ok := clientMap[userId]
	rwlocker.RUnlock()
	if ok {
		node.DataQueue <- data
	}
}
