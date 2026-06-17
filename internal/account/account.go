package account

import "gorm.io/gorm"

type Account struct {
	gorm.Model
	Username     string `gorm:"column:username;size:64;uniqueIndex;not null" json:"username"`
	PasswordHash string `gorm:"column:password_hash;size:255;not null" json:"-"`
	AvatarURL    string `gorm:"column:avatar_url;size:255" json:"avatar_url"`
	Bio          string `gorm:"column:bio;size:255" json:"bio"`
	Token        string `gorm:"column:token;type:text" json:"-"`
	RefreshToken string `gorm:"column:refresh_token;type:text" json:"-"`
}

func CreateAccount(db *gorm.DB, acc *Account) (err error) {
	err = db.Create(acc).Error
	return err
}

func UpdateAccount(db *gorm.DB, acc *Account) (err error) {
	err = db.Save(acc).Error
	return err
}

func FindAccountByID(db *gorm.DB, id uint) (acc *Account, err error) {
	acc = new(Account)
	err = db.Where("id = ?", id).First(acc).Error
	return acc, err
}

func FindAccountByName(db *gorm.DB, name string) (acc *Account, err error) {
	acc = new(Account)
	err = db.Where("username = ?", name).First(acc).Error
	return acc, err
}

func DeleteAccount(db *gorm.DB, id uint) (err error) {
	err = db.Delete(&Account{}, id).Error
	return err
}
