package logic

import (
	"net/http"
	"sync"

	"github.com/RicheyJang/key_keeper/keeper"
	"github.com/RicheyJang/key_keeper/model"
	"github.com/RicheyJang/key_keeper/utils/errors"
	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Manager 逻辑管理器：负责处理API逻辑、生成密钥保管器（实例）等
type Manager struct {
	defaultKName string   // 默认Keeper名称
	generatorMap sync.Map // keeper名称 -> 生成器(keeper.Generator)

	defaultIns  InstanceInfo // 默认实例
	instanceMap sync.Map     // 实例标识符 -> 实例信息(InstanceInfo)

	frozenUsers sync.Map           // 此次运行中被冻结的用户ID集合，用于使JWT失效
	userManager *model.UserManager // 用户管理器

	db *gorm.DB
}

// Option 创建Manager时的参数
type Option struct {
	KGs         []KeeperGeneratorPair // 首项认为是默认生成器
	DB          *gorm.DB
	UserManager *model.UserManager
}

// KeeperGeneratorPair 密钥保管器名称及其对应的生成器
type KeeperGeneratorPair struct {
	KeeperName string           // 密钥保管器名称
	Generator  keeper.Generator // keeper生成器
}

var onlyOneManager *Manager

// NewManager 创建Manager
func NewManager(option Option) (*Manager, error) {
	if onlyOneManager != nil { // 单例
		return GetManager(), nil
	}
	// option 检查
	if len(option.KGs) == 0 {
		return nil, errors.New(-1, "Initial Error: KeeperGeneratorPair is empty")
	}
	if option.DB == nil {
		return nil, errors.New(-1, "Initial Error: db is nil")
	}
	if option.UserManager == nil {
		return nil, errors.New(-1, "Initial Error: userManager is nil")
	}
	// 初始化Manager
	m := new(Manager)
	m.db = option.DB
	m.userManager = option.UserManager
	m.defaultKName = option.KGs[0].KeeperName
	for _, kg := range option.KGs {
		m.generatorMap.Store(kg.KeeperName, kg.Generator)
	}
	// 获取所有实例
	if err := m.initAllInstances(); err != nil {
		return nil, errors.Newf(-1, "InitAllInstances failed: %v", err)
	}
	defaultInstance, ok := m.getInstance(DefaultInstanceIdentifier)
	if !ok { // 至少应该有默认实例
		return nil, errors.New(-1, "LoadAllInstances failed: there is no default instance")
	}
	m.defaultIns = defaultInstance
	onlyOneManager = m
	return m, nil
}

// GetManager 获取Manager
func GetManager() *Manager {
	return onlyOneManager
}

// 回包：出错
func responseError(c iris.Context, err error) {
	errT := errors.Unknown
	if err != nil { // 构建errors.ErrorSt
		switch err.(type) {
		case errors.Error:
			errT = err.(errors.Error)
		case *errors.Error:
			errT = *(err.(*errors.Error))
		default:
			errT = errors.New(errors.CodeInner, err.Error())
		}
	}
	if errT != errors.Unknown {
		log.Errorf("API error: %v", err)
	}
	if errT.Code == errors.CodeInner {
		errT = errors.Unknown
	}
	// 回应error
	c.StopWithJSON(http.StatusInternalServerError, iris.Map{
		"code": errT.Code,
		"msg":  errT.Msg,
	})
}

// 回包：成功
func responseSuccess(c iris.Context, key string, value interface{}) {
	c.StatusCode(http.StatusOK)
	data := iris.Map{
		"code": 0,
		"msg":  "success",
	}
	if value != nil {
		data[key] = value
	}
	_, _ = c.JSON(data)
}
