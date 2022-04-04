package logic

import (
	"net/http"
	"sync"

	"github.com/RicheyJang/key_keeper/keeper"
	"github.com/RicheyJang/key_keeper/utils/errors"
	"github.com/kataras/iris/v12"
)

// Manager 逻辑管理器：负责处理API逻辑、生成密钥保管器（实例）等
type Manager struct {
	defaultKeeper keeper.KeyKeeper
	defaultKName  string   // 默认Keeper名称
	generatorMap  sync.Map // keeper名称 -> 生成器(keeper.Generator)
	keeperMap     sync.Map // 实例标识符 -> 该实例的密钥保管器
}

// Option 创建Manager时的参数
type Option struct {
	KGs []KeeperGeneratorPair
}

// KeeperGeneratorPair 密钥保管器名称及其对应的生成器
type KeeperGeneratorPair struct {
	KeeperName string           // 密钥保管器名称
	Generator  keeper.Generator // keeper生成器
	IsDefault  bool             // 是否为默认keeper
}

// DefaultKeeperOption 默认密钥保管器生成配置
var DefaultKeeperOption = keeper.Option{}

// NewManager 创建Manager
func NewManager(option Option) (*Manager, error) {
	// option 检查
	if len(option.KGs) == 0 {
		return nil, errors.New(-1, "Initial Error: KeeperGeneratorPair is empty")
	}
	m := new(Manager)
	// 创建默认keeper
	defaultG := option.KGs[0].Generator
	m.defaultKName = option.KGs[0].KeeperName
	for _, kg := range option.KGs {
		if kg.IsDefault == true {
			defaultG = kg.Generator
			m.defaultKName = kg.KeeperName
			break
		}
	}
	defaultK, err := defaultG(DefaultKeeperOption)
	if err != nil {
		return nil, err
	}
	if defaultK == nil {
		return nil, errors.New(-1, "Initial Error: Default Keeper got nil")
	}
	// 初始化Manager
	m.defaultKeeper = defaultK
	for _, kg := range option.KGs {
		m.generatorMap.Store(kg.KeeperName, kg.Generator)
	}
	return m, nil
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
	// 回应error
	c.StatusCode(http.StatusOK)
	_, _ = c.JSON(iris.Map{
		"code": errT.Code,
		"msg":  errT.Msg,
	})
}

// 回包：成功
func responseSuccess(c iris.Context, key string, value interface{}) {
	c.StatusCode(http.StatusOK)
	_, _ = c.JSON(iris.Map{
		"code": 0,
		"msg":  "success",
		key:    value,
	})
}
