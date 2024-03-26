package models

import (
	"ginchat/utils"

	"gorm.io/gorm"
)

type Contact struct {
	gorm.Model
	OwnerId  uint //所属者
	TargetId uint //联系人
	Type     int  //联系人类型
	Desc     string
}

func (table *Contact) TableName() string {
	return "contact"
}

func SearchFrend(userId uint) []UserBasic {
	contacts := make([]Contact, 0)
	objIds := make([]uint64, 0)
	utils.DB.Where("owner_id = ? and type = 1", userId).Find(&contacts)
	for _, v := range contacts {
		objIds = append(objIds, uint64(v.TargetId))
	}
	users := make([]UserBasic, 0)
	utils.DB.Where("id in (?)", objIds).Find(&users)
	return users
}
