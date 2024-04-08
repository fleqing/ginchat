package models

import (
	"context"
	"encoding/json"
	"fmt"
	"ginchat/utils"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"gopkg.in/fatih/set.v0"
	"gorm.io/gorm"
)

type Message struct {
	gorm.Model
	UserId     int64  //发送者
	TargetId   int64  //接收者
	Type       int    //发送类型 群聊、私聊、广播
	Media      int    //消息类型 文本、图片、视频、文件
	Content    string //消息内容
	CreateTime uint64 // 创建时间
	ReadTime   uint64 // 读取时间
	Pic        string
	Url        string
	Desc       string
	Amount     int //其他数字统计
}

func (table *Message) TableName() string {
	return "message"
}

type Node struct {
	Conn          *websocket.Conn
	Addr          string //客户端地址
	FirstTime     uint64 //首次连接时间
	HeartbeatTime uint64 //心跳时间
	LoginTime     uint64 //登录时间
	DataQueue     chan []byte
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
	currentTime := uint64(time.Now().Unix())
	node := &Node{Conn: conn,
		Addr:          conn.RemoteAddr().String(), //客户端地址
		HeartbeatTime: currentTime,                //心跳时间
		LoginTime:     currentTime,                //登录时间
		DataQueue:     make(chan []byte, 50),
		GroupSets:     set.New(set.ThreadSafe)}
	// 将用户的连接信息存储到map中
	rwlocker.Lock()
	clientMap[userId] = node
	rwlocker.Unlock()
	// 开启协程，进行消息的发送
	go sendProc(node)
	// 开启协程，进行消息的接收
	go recvProc(node)
	// 将用户的在线信息存储到redis中
	SetUserOnlineInfo("online_"+Id, []byte(node.Addr), time.Duration(viper.GetInt("timeout.RedisOnlineTime"))*time.Hour)
	// sendMsg(userId, []byte("欢迎进入聊天系统"))
}

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
		fmt.Println("udp err ->>>>", err)
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
		fmt.Println("udp err ->>>>", err)
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
	// 群聊
	case 2:
		sendGroupMsg(msg.TargetId, data)
	}

}

// 将数据放在对应的用户的 DataQueue 中
func sendMsg(userId int64, data []byte) {
	fmt.Println("sendmsg >>> userid", userId, "msg", string(data))
	rwlocker.RLock()
	node, ok := clientMap[userId]
	rwlocker.RUnlock()
	jsonMsg := Message{}
	json.Unmarshal(data, &jsonMsg)
	ctx := context.Background()
	targetIdStr := strconv.Itoa(int(userId))
	// 博主后期改了代码，将FromId改为了UserId
	userIdStr := strconv.Itoa(int(jsonMsg.UserId))
	jsonMsg.CreateTime = uint64(time.Now().Unix())
	r, err := utils.Red.Get(ctx, "online_"+userIdStr).Result()
	if err != nil {
		fmt.Println(err)
	}
	// 检查用户是否在线，如果在线，就将数据发送给用户
	if r != "" {
		if ok {
			fmt.Println("sendMsg >>> userId", userId, "data:", string(data))
			node.DataQueue <- data
		}
	}
	var key string
	if userId > jsonMsg.UserId {
		key = "msg_" + userIdStr + "_" + targetIdStr
	} else {
		key = "msg_" + targetIdStr + "_" + userIdStr
	}
	res, err := utils.Red.ZRevRange(ctx, key, 0, -1).Result()
	if err != nil {
		fmt.Println(err)
	}
	// 两个用户之间的消息都有一个分数，分数越大，前端的消息越靠后，和qq差不多
	score := float64(cap(res)) + 1
	ress, e := utils.Red.ZAdd(ctx, key, &redis.Z{Score: score, Member: data}).Result()
	if e != nil {
		fmt.Println(e)
	}
	// 如果成功，打印被成功添加的元素数量。
	fmt.Println("元素的数量", ress)
}

func sendGroupMsg(targetId int64, msg []byte) {
	fmt.Println("开始群发消息")
	userIds := SearchUserByGroupId(uint(targetId))
	for i := 0; i < len(userIds); i++ {
		if targetId != int64(userIds[i]) {
			sendMsg(int64(userIds[i]), msg)
		}
	}
}

// 获取缓存里面的消息
func RedisMsg(userIdA int64, userIdB int64, start int64, end int64, isRev bool) []string {
	rwlocker.RLock()
	rwlocker.RUnlock()
	// Context 是用于在不同的 goroutine 之间传递 deadline、取消信号和其他请求范围的值的机制。
	ctx := context.Background()
	userIdStr := strconv.Itoa(int(userIdA))
	targetIdStr := strconv.Itoa(int(userIdB))
	var key string
	if userIdA > userIdB {
		key = "msg_" + targetIdStr + "_" + userIdStr
	} else {
		key = "msg_" + userIdStr + "_" + targetIdStr
	}
	var rels []string
	var err error
	if isRev {
		rels, err = utils.Red.ZRange(ctx, key, start, end).Result()
	} else {
		rels, err = utils.Red.ZRevRange(ctx, key, start, end).Result()
	}
	if err != nil {
		fmt.Println(err)
	}
	return rels
}
