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

func AddFriend(userId uint, targetName string) (int, string) {
	user := FindUserByName(targetName)
	if user.Identity != "" {

		if userId == user.ID {
			return -1, "不能添加自己为好友"
		}
		contact0 := Contact{}
		utils.DB.Where("owner_id = ? and target_id = ? and type = 1", userId, user.ID).First(&contact0)
		if contact0.ID != 0 {
			return -1, "已经是好友"
		}
		tx := utils.DB.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		contact := Contact{
			OwnerId:  userId,
			TargetId: user.ID,
			Type:     1,
		}
		if err := utils.DB.Create(&contact).Error; err != nil {
			tx.Rollback()
			return -1, "添加好友失败"
		}
		contact2 := Contact{
			OwnerId:  user.ID,
			TargetId: userId,
			Type:     1,
		}
		if err := utils.DB.Create(&contact2).Error; err != nil {
			tx.Rollback()
			return -1, "添加好友失败"
		}
		tx.Commit()
		return 0, "添加好友成功"
	}
	return -1, "用户不存在"
}

func SearchUserByGroupId(groupId uint) []uint {
	contacts := make([]Contact, 0)
	objIds := make([]uint, 0)
	utils.DB.Where("target_id = ? and type = 2", groupId).Find(&contacts)
	for _, v := range contacts {
		objIds = append(objIds, v.OwnerId)
	}
	return objIds
}
