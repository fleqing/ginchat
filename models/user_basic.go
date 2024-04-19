package models

import (
	"fmt"
	"ginchat/utils"

	"time"

	"gorm.io/gorm"
)

/*
*
`gorm:"default:CURRENT_TIMESTAMP”`
`gorm:"column:login_out_time;default:CURRENT_TIMESTAMP" json:"login_out_time"`
*
*/
type UserBasic struct {
	gorm.Model
	Name         string
	PassWord     string
	Phone        string `valid:"matches(^1[3-9]\\d{9}$)"`
	Email        string `valid:"email"`
	Identity     string
	ClientIp     string
	ClientPort   string
	Salt         string
	LoginTime    time.Time `gorm:"default:CURRENT_TIMESTAMP(3)"`
	HearBeatTime time.Time `gorm:"default:CURRENT_TIMESTAMP(3)"`
	LoginOutTime time.Time `gorm:"default:CURRENT_TIMESTAMP(3)"`
	IsLogout     bool
	DeviceInfo   string
	Icon         string
}

func (table *UserBasic) TableName() string {
	return "user_basic"
}

func GetUserList() []*UserBasic {
	data := make([]*UserBasic, 10)
	utils.DB.Find(&data)
	for _, v := range data {
		fmt.Println("data", v)
	}
	return data
}

func FindUserByNameAndPassWord(name string, password string) UserBasic {
	user := &UserBasic{}
	utils.DB.Where("name = ? and pass_word = ?", name, password).First(user)

	str := fmt.Sprintf("%d", time.Now().Unix())
	temp := utils.MD5Encode(str)
	utils.DB.Model(&user).Where("id = ?", user.ID).Update("identity", temp)
	return *user
}

func FindUserByName(name string) UserBasic {
	user := &UserBasic{}
	utils.DB.Where("name = ?", name).First(user)
	return *user
}

func FindUserByPhone(phone string) *gorm.DB {
	user := &UserBasic{}
	return utils.DB.Where("phone = ?", phone).First(user)
}

func FindUserByEmail(email string) *gorm.DB {
	user := &UserBasic{}
	return utils.DB.Where("email = ?", email).First(user)
}

// 如果你的函数不返回 *gorm.DB，那么调用者就无法知道操作是否成功，或者在出错时无法获取到错误信息。
func CreateUser(user UserBasic) *gorm.DB {
	return utils.DB.Create(&user)
}

func DeleteUser(user UserBasic) *gorm.DB {
	return utils.DB.Delete(&user)
}

// 不能直接使用 utils.DB.Updates(UserBasic{Name: user.Name, PassWord: user.PassWord})
// 这样的方式来更新数据。因为在这种情况下，GORM 不知道你想要更新哪个表和哪个记录。
func UpdateUser(user UserBasic) *gorm.DB {
	return utils.DB.Model(&user).Updates(UserBasic{Name: user.Name, PassWord: user.PassWord, Phone: user.Phone, Email: user.Email, Icon: user.Icon})
}

func FindUserById(id uint) UserBasic {
	user := &UserBasic{}
	utils.DB.Where("id = ?", id).First(user)
	return *user
}
