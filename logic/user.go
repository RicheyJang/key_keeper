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

type ChangePasswdRequest struct {
	UserID      uint   `json:"id"`          // 为空代表修改自己的密码；否则代表修改指定用户的密码
	OldPassword string `json:"oldPassword"` // 修改指定用户密码时，此项可以为空
	NewPassword string `json:"newPassword"`
}

// HandlerOfChangePasswd 修改密码处理函数
func (manager *Manager) HandlerOfChangePasswd(ctx iris.Context) {
	// 校验参数
	var request ChangePasswdRequest
	err := ctx.ReadJSON(&request)
	if err != nil {
		responseError(ctx, errors.InvalidRequest)
		return
	}
	// 权限检查
	self := manager.getUserClaims(ctx)
	if request.UserID != 0 && request.UserID != self.ID { // 修改其他用户的密码
		if self.Level < model.UserLevelRoot {
			responseError(ctx, errors.PermissionDeny)
			return
		}
	} else { // 自己修改自己的密码
		if _, err = manager.userManager.CheckUser(self.Name, request.OldPassword); err != nil {
			responseError(ctx, errors.WrongPasswd)
			return
		}
		request.UserID = self.ID
	}
	// 修改密码
	if err = manager.userManager.ChangePasswd(request.UserID, request.NewPassword); err != nil {
		responseError(ctx, err)
		return
	}
	responseSuccess(ctx, "", nil)
}
