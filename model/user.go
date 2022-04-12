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
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"column:name;uniqueIndex" json:"name"`
	Passwd    string    `gorm:"column:passwd" json:"-"`
	Level     int       `gorm:"column:level" json:"level"`
	IsFrozen  bool      `gorm:"column:is_frozen" json:"isFrozen"`
	LastLogin time.Time `gorm:"column:last_login" json:"lastLogin"`
	LastIP    string    `gorm:"column:last_ip" json:"lastIP"`
	CreatedAt time.Time `json:"createTime"`
	UpdatedAt time.Time `json:"updatedAt"`
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

// Get 获取用户
func (m *UserManager) Get(id uint) (*User, error) {
	var user User
	if err := m.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Add 新增用户
func (m *UserManager) Add(user *User) error {
	// 验证用户
	if user == nil || user.ID != 0 || len(user.Name) == 0 || len(user.Passwd) == 0 {
		return errors.InvalidRequest
	}
	var count int64
	m.db.Model(&User{}).Where("name = ?", user.Name).Count(&count)
	if count > 0 { // 用户名已存在
		return errors.UserExist
	}
	// 新增用户
	user.Passwd = passwdToSha256(user.Passwd)
	return m.db.Create(user).Error
}

type UserFilter struct {
	Offset int
	Limit  int
}

// Filter 根据UserFilter过滤用户
func (m *UserManager) Filter(filter UserFilter) ([]User, error) {
	var users []User
	if err := m.setupFilterSession(filter).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// Count 根据UserFilter查询符合条件的用户数量
func (m *UserManager) Count(filter UserFilter) int64 {
	var count int64
	if err := m.setupFilterSession(filter).Count(&count).Error; err != nil {
		return 0
	}
	return count
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
	return m.db.Model(&user).UpdateColumns(User{ // 不更新updated_at
		LastLogin: user.LastLogin,
		LastIP:    user.LastIP,
	}).Error
}

// Freeze 冻结用户
func (m *UserManager) Freeze(id uint, isFrozen bool) error {
	return m.db.Model(&User{}).Where("id = ?", id).Update("is_frozen", isFrozen).Error
}

// Delete 删除用户
func (m *UserManager) Delete(id uint) error {
	if id == 0 {
		return errors.InvalidRequest
	}
	return m.db.Delete(&User{}, id).Error
}

// SetLevel 设置用户权限等级
func (m *UserManager) SetLevel(id uint, level int) error {
	if id == 0 {
		return errors.InvalidRequest
	}
	return m.db.Model(&User{}).Where("id = ?", id).Update("level", level).Error
}

// ChangePasswd 修改用户密码
func (m *UserManager) ChangePasswd(id uint, passwd string) error {
	if len(passwd) == 0 {
		return errors.InvalidRequest
	}
	return m.db.Model(&User{}).Where("id = ?", id).Update("passwd", passwdToSha256(passwd)).Error
}

func (m *UserManager) setupFilterSession(filter UserFilter) *gorm.DB {
	db := m.db.Model(&User{})
	if filter.Offset > 0 {
		db = db.Offset(filter.Offset)
	}
	if filter.Limit > 0 {
		db = db.Limit(filter.Limit)
	}
	// 若没有排序要求，则按ID排序
	db = db.Order("id")
	return db
}

func passwdToSha256(passwd string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(passwd)))
}
