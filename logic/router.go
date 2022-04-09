package logic

import (
	"github.com/RicheyJang/key_keeper/keeper"
	"github.com/kataras/iris/v12"
)

const ctxKeeperKey = "key-keeper"

// PreRouterOfSetKeeper Router中间件：根据请求实例查询处理该请求的密钥保管器
func (manager *Manager) PreRouterOfSetKeeper(ctx iris.Context) {
	// TODO 分实例给予不同的keeper
	ctx.Values().Set(ctxKeeperKey, manager.defaultIns.kp)
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
