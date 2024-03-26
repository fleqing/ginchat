package models

import "gorm.io/gorm"

// GroupBasic 群组基础信息
type GroupBasic struct {
	gorm.Model
	Name    string //群名称
	OwnerId uint   //群主
	Icon    string //群头像
	Type    int    //群类型
	Desc    string //群描述
}

func (table *GroupBasic) TableName() string {
	return "group_basic"
}
