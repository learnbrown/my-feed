// 业务逻辑
package account

import (
	"errors"
	"my_feed/internal/auth"
	"my_feed/internal/db"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type AccountService struct {
	repo *AccountRepo
}

func NewAccountService(repo *AccountRepo) *AccountService {
	return &AccountService{repo: repo}
}

// 定义错误类型
var (
	ErrUsernameRequired          = errors.New("username required")
	ErrUsernameExists            = errors.New("username exists")
	ErrPasswordRequired          = errors.New("password required")
	ErrInvalidUsernameOrPassword = errors.New("invalid username or password")
	ErrAccountNotFound           = errors.New("account not found")
)

func (service *AccountService) Register(username string, password string) (*Account, error) {
	// 去除用户名中的空格
	username = strings.TrimSpace(username)

	// 检查用户名是否为空
	if username == "" {
		return nil, ErrUsernameRequired
	}

	// 检查密码是否为空
	if password == "" {
		return nil, ErrPasswordRequired
	}

	// 查询用户名是否存在
	exists, err := service.repo.ExistAccount(username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUsernameExists
	}

	// 创建用户记录
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	acc := &Account{
		Username:     username,
		PasswordHash: string(passwordHash),
	}

	err = service.repo.CreateAccount(acc)
	if err != nil {
		// TODO
		// 并发场景下可能会触发唯一索引错误
		// 查重后有用户注册
		if db.IsDuplicateKeyError(err) {
			return nil, ErrUsernameExists
		}
		return nil, err
	}

	return acc, nil
}

func (service *AccountService) Login(username string, password string) (*Account, error) {
	// 查询用户是否存在
	username = strings.TrimSpace(username)

	if username == "" || password == "" {
		return nil, ErrInvalidUsernameOrPassword
	}

	acc, err := service.repo.FindAccountByName(username)
	if err != nil {
		// 用户不存在或密码错误返回相同的信息，避免泄漏用户名是否存在的信息
		// 将未查到用户和数据库错误的处理分开
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, ErrInvalidUsernameOrPassword
		}
		return nil, err
	}

	// 比对登录信息
	err = bcrypt.CompareHashAndPassword([]byte(acc.PasswordHash), []byte(password))
	if err != nil {
		return nil, ErrInvalidUsernameOrPassword
	}

	// 生成JWTtoken
	token, err := auth.GenerateToken(acc.ID, acc.Username)
	if err != nil {
		return nil, err
	}

	// 更新用户记录
	err = service.repo.UpdateToken(acc.ID, token)
	if err != nil {
		return nil, err
	}

	acc.Token = token

	return acc, nil
}

func (service *AccountService) Logout(id uint) error {
	err := service.repo.UpdateToken(id, "")
	if errors.Is(err, db.ErrRecordNotFound) {
		return ErrAccountNotFound
	}
	return err
}

func (service *AccountService) Me(id uint) (*Account, error) {
	acc, err := service.repo.FindAccountByID(id)
	if errors.Is(err, db.ErrRecordNotFound) {
		return nil, ErrAccountNotFound
	}
	return acc, err
}
