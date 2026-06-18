package account

import "gorm.io/gorm"

type Account struct {
	gorm.Model
	Username     string `gorm:"column:username;size:64;uniqueIndex;not null" json:"username"`
	PasswordHash string `gorm:"column:password_hash;size:255;not null" json:"-"`
	// 用户头像
	AvatarURL string `gorm:"column:avatar_url;size:255" json:"avatar_url"`
	// 用户简介
	Bio string `gorm:"column:bio;size:255" json:"bio"`

	Token        string `gorm:"column:token;type:text" json:"-"`
	RefreshToken string `gorm:"column:refresh_token;type:text" json:"-"`
}
