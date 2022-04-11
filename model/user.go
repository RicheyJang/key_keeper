package model

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/RicheyJang/key_keeper/utils/errors"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	UserLevelGeneral = iota
	UserLevelAdmin
	UserLevelRoot
)

type User struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"column:name;uniqueIndex"`
	Passwd    string    `gorm:"column:passwd"`
	Level     int       `gorm:"column:level"`
	IsFrozen  bool      `gorm:"column:is_frozen"`
	LastLogin time.Time `gorm:"column:last_login"`
	LastIP    string    `gorm:"column:last_ip"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (user User) TableName() string {
	return "t_manager_users"
}

type UserManager struct {
	db *gorm.DB
}

// NewUserManger 创建新的用户管理器
func NewUserManger(db *gorm.DB) *UserManager {
	if db == nil {
		return nil
	}
	// 初始化数据库
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Errorf("NewUserManger error: %v", err)
	}
	var count int64
	db.Model(&User{}).Count(&count)
	if count == 0 {
		// 初始化root
		user := User{
			Name:     "root",
			Passwd:   passwdToSha256("root"),
			Level:    UserLevelRoot,
			IsFrozen: false,
		}
		db.Create(&user)
	}
	// 初始化用户管理器
	m := &UserManager{
		db: db,
	}
	return m
}

// GetUser 获取用户
func (m *UserManager) GetUser(id uint) (*User, error) {
	var user User
	if err := m.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// CheckUser 验证用户名密码
func (m *UserManager) CheckUser(name, passwd string) (User, error) {
	var user User
	if err := m.db.Where("name = ?", name).First(&user).Error; err != nil {
		return user, err
	}
	if passwdToSha256(passwd) != user.Passwd {
		return user, errors.WrongPasswd
	}
	return user, nil
}

// SaveUserLoginInfo 更新用户登录信息
func (m *UserManager) SaveUserLoginInfo(user User) error {
	if user.ID == 0 {
		return errors.InvalidRequest
	}
	return m.db.Model(&user).Updates(User{
		LastLogin: user.LastLogin,
		LastIP:    user.LastIP,
	}).Error
}

// FreezeUser 冻结用户
func (m *UserManager) FreezeUser(id uint, isFrozen bool) error {
	return m.db.Model(&User{}).Where("id = ?", id).Update("is_frozen", isFrozen).Error
}

func passwdToSha256(passwd string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(passwd)))
}
