package models

import (
	"ginchat/utils"

	"gorm.io/gorm"
)

type Community struct {
	gorm.Model
	Name    string
	OwnerId uint
	Img     string
	Desc    string
}

func CreateCommunity(community Community) (int, string) {
	if community.Name == "" {
		return -1, "群名不能为空"
	}
	if community.OwnerId == 0 {
		return -1, "请先登录"
	}
	if err := utils.DB.Create(&community).Error; err != nil {
		return -1, "创建群失败"
	}
	return 0, "创建群成功"
}

func LoadCommunity(userId uint) []Community {
	var contact []Contact
	var objIds []uint
	utils.DB.Where("owner_id = ? and type = 2", userId).Find(&contact)
	for _, v := range contact {
		objIds = append(objIds, v.TargetId)
	}
	var community []Community
	utils.DB.Where("id in (?)", objIds).Find(&community)
	return community
}

func JoinGroup(userId uint, communityId string) (int, string) {
	var community Community
	utils.DB.Where("id = ? or name = ?", communityId, communityId).First(&community)
	if community.ID == 0 {
		return -1, "群不存在"
	}
	contact := Contact{}
	contact.OwnerId = userId
	contact.Type = 2
	utils.DB.Where("owner_id = ? and target_id = ? and type = 2", userId, community.ID).First(&contact)
	if contact.ID != 0 {
		return -1, "已经加入群聊"
	} else {
		contact.TargetId = community.ID
		if err := utils.DB.Create(&contact).Error; err != nil {
			return -1, "加入群聊失败"
		}
	}
	return 0, "加入群聊成功"
}
