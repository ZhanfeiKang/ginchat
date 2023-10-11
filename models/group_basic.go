package models

import "gorm.io/gorm"

// 群信息
type GroupBasic struct {
	gorm.Model
	Name    string
	OwnerId uint   // 群主
	Icon    string // 群的图片
	Type    int    // 预留，充了钱提升等级
	Desc    string
}

func (table *GroupBasic) TableName() string {
	return "group_basic"
}
