package logic

import (
	"time"

	"github.com/RicheyJang/key_keeper/model"
	"github.com/RicheyJang/key_keeper/utils/errors"
	"github.com/kataras/iris/v12"
)

type FreezeUserRequest struct {
	UserID   uint `json:"userID"`
	IsFrozen bool `json:"isFrozen"`
}

// HandlerOfFreezeUser 冻结或解冻用户处理函数
func (manager *Manager) HandlerOfFreezeUser(ctx iris.Context) {
	// 权限检查
	self := manager.getUserClaims(ctx)
	if self.Level < model.UserLevelAdmin {
		responseError(ctx, errors.PermissionDeny)
		return
	}
	// 校验参数
	var request FreezeUserRequest
	err := ctx.ReadJSON(&request)
	if err != nil {
		responseError(ctx, errors.InvalidRequest)
		return
	}
	if request.UserID == self.ID {
		responseError(ctx, errors.InvalidRequest)
		return
	}
	// TODO 冻结该用户的所有实例
	// 保证被冻结用户的权限级别低于调用者
	user, err := manager.userManager.GetUser(request.UserID)
	if err != nil {
		responseError(ctx, err)
		return
	}
	if user.Level >= self.Level {
		responseError(ctx, errors.PermissionDeny)
		return
	}
	// 冻结或解冻
	err = manager.userManager.FreezeUser(request.UserID, request.IsFrozen)
	if err != nil {
		responseError(ctx, err)
		return
	}
	if request.IsFrozen {
		manager.frozenUsers.Store(request.UserID, time.Now())
	} else {
		manager.frozenUsers.Delete(request.UserID)
	}
	responseSuccess(ctx, "", nil)
}
