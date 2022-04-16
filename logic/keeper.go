package logic

import (
	"github.com/RicheyJang/key_keeper/keeper"
	"github.com/RicheyJang/key_keeper/utils/errors"
	"github.com/kataras/iris/v12"
)

const ctxKeeperKey = "key-keeper"

// PreRouterOfSetKeeper Router中间件：根据请求实例查询处理该请求的密钥保管器
func (manager *Manager) PreRouterOfSetKeeper(ctx iris.Context) {
	// 分实例给予不同的keeper
	identifier := ctx.GetHeader("identifier")
	info, ok := manager.getInstance(identifier)
	if !ok { // 不存在该实例
		responseError(ctx, errors.NoSuchInstance)
		return
	}
	if info.IsFrozen { // 实例被冻结
		responseError(ctx, errors.InstanceFrozen)
		return
	}
	ctx.Values().Set(ctxKeeperKey, info.kp)
	ctx.Next()
}

// 获取当前上下文环境下的密钥保管器
func (manager *Manager) getKeeper(ctx iris.Context) keeper.KeyKeeper {
	k, ok := ctx.Values().Get(ctxKeeperKey).(keeper.KeyKeeper)
	if !ok || k == nil {
		return manager.defaultIns.kp
	}
	return k
}
