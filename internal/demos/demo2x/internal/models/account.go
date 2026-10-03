package models

import "gorm.io/gorm"

type AccountV1 struct {
	gorm.Model
	Username string `gorm:"unique"`
	Nickname string `gorm:"column:nickname"`
	Rank     string `gorm:"column:rank"`
	Score    string `gorm:"column:score"`
}

func (*AccountV1) TableName() string {
	return "users"
}

type AccountV2 struct {
	gorm.Model
	Username string `gorm:"unique"`
	Nickname string `gorm:"column:nickname"`
	Rank     uint64 `gorm:"column:rank"`
	Score    string `gorm:"column:score"`
}

func (*AccountV2) TableName() string {
	return "users"
}

type AccountV3 struct {
	gorm.Model
	Username string  `gorm:"unique"`
	Nickname string  `gorm:"column:nickname"`
	Rank     uint64  `gorm:"column:rank"`
	Score    float64 `gorm:"column:score"`
}

func (*AccountV3) TableName() string {
	return "users"
}
