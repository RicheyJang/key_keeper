package logic

import (
	"github.com/RicheyJang/key_keeper/keeper"
	"github.com/kataras/iris/v12"
)

// GetKeyInfo 获取指定密钥
func (manager *Manager) GetKeyInfo(ctx iris.Context) {
	k := manager.getKeeper(ctx)
	// 解析参数
	var req keeper.KeyRequest
	if err := ctx.ReadJSON(&req); err != nil {
		responseError(ctx, err)
		return
	}
	// 获取密钥信息
	key, err := k.GetKeyInfo(req)
	if err != nil {
		responseError(ctx, err)
	} else {
		responseSuccess(ctx, "key", key)
	}
}

// GetLatestVersionKey 获取特定密钥ID下的最新版本密钥
func (manager *Manager) GetLatestVersionKey(ctx iris.Context) {
	k := manager.getKeeper(ctx)
	// 解析参数
	var req keeper.KeyRequest
	if err := ctx.ReadJSON(&req); err != nil {
		responseError(ctx, err)
		return
	}
	// 获取密钥最新版本
	key, err := k.GetLatestVersionKey(req.ID)
	if err != nil {
		responseError(ctx, err)
	} else {
		responseSuccess(ctx, "key", key)
	}
}
