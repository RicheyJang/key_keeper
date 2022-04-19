package logic

import (
	"math"
	"strings"

	"github.com/RicheyJang/key_keeper/keeper"
	"github.com/RicheyJang/key_keeper/utils/errors"
	"github.com/kataras/iris/v12"
)

const ctxUserInstanceKey = "user-instance"

// PreCheckOfUserInstance 检查当前用户对所用实例的权限
func (manager *Manager) PreCheckOfUserInstance(ctx iris.Context) {
	identifier := ctx.GetHeader("identifier") // 从Header中获取实例标识
	i, err := manager.getInstanceAndCheckUser(identifier, ctx)
	if err != nil {
		responseError(ctx, err)
		return
	}
	if i == nil {
		responseError(ctx, errors.NoSuchInstance)
		return
	}
	if i.IsFrozen {
		responseError(ctx, errors.InstanceFrozen)
		return
	}
	ctx.Values().Set(ctxUserInstanceKey, i)
	ctx.Next()
}

type GetKeysRequest struct {
	Page int `url:"page"`
	Size int `url:"size"`
}

// HandlerOfGetKeys 批量获取密钥信息
func (manager *Manager) HandlerOfGetKeys(ctx iris.Context) {
	// 解析请求
	var request GetKeysRequest
	err := ctx.ReadQuery(&request)
	if err != nil {
		responseError(ctx, err)
		return
	}
	if request.Page < 0 || request.Size < 0 {
		responseError(ctx, errors.InvalidRequest)
		return
	}
	instance := manager.getUserInstance(ctx)
	// 获取密钥列表
	keys, total, err := instance.kp.FilterKeys(keeper.KeysFilter{
		Offset:  request.Size * (request.Page - 1),
		Limit:   request.Size,
		Content: false,
	})
	if err != nil {
		responseError(ctx, err)
		return
	}
	for i := range keys { // 保证密钥内容不暴露
		keys[i].Key = ""
	}
	// 返回结果
	if len(keys) == 0 {
		keys = make([]keeper.KeyInfo, 0)
	}
	responseSuccess(ctx, "data", iris.Map{
		"keys":  keys,
		"total": total,
	})
}

// HandlerOfAddKey 派发新密钥
func (manager *Manager) HandlerOfAddKey(ctx iris.Context) {
	// 解析请求
	var request keeper.DistributeKeyRequest
	err := ctx.ReadJSON(&request)
	if err != nil {
		responseError(ctx, err)
		return
	}
	// 请求校验
	if request.ID < 1 || request.ID > math.MaxUint32 ||
		!(request.Length == 16 || request.Length == 24 || request.Length == 32) ||
		!strings.HasPrefix(request.Algorithm, "aes") {
		responseError(ctx, errors.InvalidRequest)
		return
	}
	instance := manager.getUserInstance(ctx)
	// 派发密钥
	key, err := instance.kp.DistributeKey(request)
	if err != nil {
		responseError(ctx, err)
		return
	}
	// 返回结果
	key.Key = ""
	responseSuccess(ctx, "data", iris.Map{
		"key": key,
	})
}

// HandlerOfDestroyKey 销毁密钥处理函数
func (manager *Manager) HandlerOfDestroyKey(ctx iris.Context) {
	// 解析请求
	id := ctx.URLParamUint64("id")
	if id == 0 {
		responseError(ctx, errors.InvalidRequest)
		return
	}
	instance := manager.getUserInstance(ctx)
	// 销毁密钥
	err := instance.kp.DestroyKey(uint(id))
	if err != nil {
		responseError(ctx, err)
		return
	}
	// 返回结果
	responseSuccess(ctx, "", nil)
}

// 仅可用于web的/keys系列API，即有PreCheckOfUserInstance中间件
func (manager *Manager) getUserInstance(ctx iris.Context) *InstanceInfo {
	return ctx.Values().Get(ctxUserInstanceKey).(*InstanceInfo)
}
