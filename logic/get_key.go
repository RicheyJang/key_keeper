package logic

import (
	"github.com/RicheyJang/key_keeper/keeper"
	"github.com/kataras/iris/v12"
)

func (manager *Manager) GetKeyInfo(ctx iris.Context) {
	k := manager.getKeeper(ctx)
	// TODO 定义API 补全Request
	key, err := k.GetKeyInfo(keeper.KeyRequest{
		ID:      1,
		Version: 1,
	})
	if err != nil {
		responseError(ctx, err)
	} else {
		responseSuccess(ctx, "key", key)
	}
}
