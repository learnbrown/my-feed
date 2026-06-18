// DB操作
package account

import (
	"my_feed/internal/db"

	"gorm.io/gorm"
)

// 对象式repository
type AccountRepo struct {
	db *gorm.DB
}

func NewAccountRepo(db *gorm.DB) *AccountRepo {
	return &AccountRepo{db: db}
}

func (this *AccountRepo) CreateAccount(acc *Account) (err error) {
	err = this.db.Create(acc).Error
	return err
}

func (this *AccountRepo) UpdateToken(id uint, token string) (err error) {
	result := this.db.Model(&Account{}).
		Where("id = ?", id).
		Update("token", token)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return db.ErrRecordNotFound
	}

	return nil
}

func (this *AccountRepo) FindAccountByID(id uint) (acc *Account, err error) {
	acc = new(Account)
	err = this.db.Where("id = ?", id).First(acc).Error
	return acc, err
}

func (this *AccountRepo) FindAccountByName(name string) (acc *Account, err error) {
	acc = new(Account)
	err = this.db.Where("username = ?", name).First(acc).Error
	return acc, err
}

func (this *AccountRepo) DeleteAccount(id uint) (err error) {
	err = this.db.Delete(&Account{}, id).Error
	return err
}

func (this *AccountRepo) ExistAccount(username string) (exists bool, err error) {
	var count int64
	err = this.db.Model(&Account{}).
		Where("username = ?", username).
		Count(&count).
		Error
	return count > 0, err
}
