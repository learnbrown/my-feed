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

func (repo *AccountRepo) CreateAccount(acc *Account) error {
	err := repo.db.Create(acc).Error
	return err
}

func (repo *AccountRepo) UpdateToken(id uint, token string) error {
	result := repo.db.Model(&Account{}).
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

func (repo *AccountRepo) FindAccountByID(id uint) (*Account, error) {
	acc := &Account{}
	err := repo.db.Where("id = ?", id).First(acc).Error
	return acc, err
}

func (repo *AccountRepo) FindAccountByName(name string) (*Account, error) {
	acc := &Account{}
	err := repo.db.Where("username = ?", name).First(acc).Error
	return acc, err
}

func (repo *AccountRepo) DeleteAccount(id uint) error {
	err := repo.db.Delete(&Account{}, id).Error
	return err
}

func (repo *AccountRepo) ExistAccount(username string) (bool, error) {
	var count int64
	err := repo.db.Model(&Account{}).
		Where("username = ?", username).
		Count(&count).
		Error
	return count > 0, err
}
