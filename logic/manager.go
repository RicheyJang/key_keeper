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
	defaultIns   InstanceInfo // 默认实例
	defaultKName string       // 默认Keeper名称
	generatorMap sync.Map     // keeper名称 -> 生成器(keeper.Generator)
	instanceMap  sync.Map     // 实例标识符 -> 实例信息(InstanceInfo)
}

// Option 创建Manager时的参数
type Option struct {
	KGs []KeeperGeneratorPair // 首项认为是默认生成器
}

// KeeperGeneratorPair 密钥保管器名称及其对应的生成器
type KeeperGeneratorPair struct {
	KeeperName string           // 密钥保管器名称
	Generator  keeper.Generator // keeper生成器
}

// NewManager 创建Manager
func NewManager(option Option) (*Manager, error) {
	// option 检查
	if len(option.KGs) == 0 {
		return nil, errors.New(-1, "Initial Error: KeeperGeneratorPair is empty")
	}
	// 获取所有实例
	instances, err := LoadAllInstances(option.KGs)
	if err != nil {
		return nil, errors.Newf(-1, "LoadAllInstances failed: %v", err)
	}
	if len(instances) == 0 || instances[0].Identifier != DefaultInstanceID { // 至少应该有默认实例
		return nil, errors.New(-1, "LoadAllInstances failed: there is no default instance")
	}
	// 初始化Manager
	m := new(Manager)
	m.defaultKName = option.KGs[0].KeeperName
	m.defaultIns = instances[0]
	for _, kg := range option.KGs {
		m.generatorMap.Store(kg.KeeperName, kg.Generator)
	}
	for _, ins := range instances {
		m.instanceMap.Store(ins.Identifier, ins)
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
	c.StatusCode(http.StatusInternalServerError)
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
