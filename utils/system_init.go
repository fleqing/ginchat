package utils

import (
	"context"
	"fmt"

	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func init() {
	InitConfig()
	InitMysql()
	InitRedis()
}

func InitConfig() {
	//TODO
	viper.SetConfigName("app")
	viper.AddConfigPath("config")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
	fmt.Println("config mysql", viper.Get("mysql"))
	fmt.Println("config app", viper.Get("app"))

}

var (
	DB  *gorm.DB      //数据库连接
	Red *redis.Client //redis连接
)

func InitMysql() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info, // Log level
			Colorful:      true,        // Disable color
		},
	)
	DB, _ = gorm.Open(mysql.Open(viper.GetString("mysql.dns")), &gorm.Config{Logger: newLogger})
	// DB.AutoMigrate(&models.UserBasic{})
	// user := &models.UserBasic{}
	// DB.Find(&user)
	// fmt.Println("user", user)

}

func InitRedis() {
	//TODO
	Red = redis.NewClient(&redis.Options{
		Addr:         viper.GetString("redis.addr"),
		Password:     viper.GetString("redis.password"), // no password set
		DB:           viper.GetInt("redis.db"),          // use default DB
		PoolSize:     viper.GetInt("redis.pool_size"),
		MinIdleConns: viper.GetInt("redis.min_idle_conns"),
	})
	pong, err := Red.Ping(context.Background()).Result()
	if err != nil {
		fmt.Println("redis connect error", err)
	} else {
		fmt.Println("redis connect success", pong)
	}
}

const (
	PublishKey = "websocket"
)

// 发布消息到redis
func Publish(ctx context.Context, channel string, msg string) error {
	err := Red.Publish(ctx, channel, msg).Err()
	if err != nil {
		fmt.Println("err", err)
	}
	return err
}

// 订阅消息
func Subscribe(ctx context.Context, channel string) (string, error) {
	// 这个对象可以用来接收通道的消息
	sub := Red.Subscribe(ctx, channel)
	fmt.Println("sub", ctx)
	msg, err := sub.ReceiveMessage(ctx)
	if err != nil {
		fmt.Println("err", err)
		return "", err
	}
	fmt.Println("msg", msg)
	return msg.Payload, err
}
