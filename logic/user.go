package logic

import (
	"time"

	"github.com/RicheyJang/key_keeper/model"
	"github.com/RicheyJang/key_keeper/utils/errors"
	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
)

type GetUsersRequest struct {
	Page int `url:"page"`
	Size int `url:"size"`
}

// HandlerOfGetUsers 批量获取用户处理函数
func (manager *Manager) HandlerOfGetUsers(ctx iris.Context) {
	// 权限检查
	self := manager.getUserClaims(ctx)
	if self.Level < model.UserLevelAdmin {
		responseError(ctx, errors.PermissionDeny)
		return
	}
	// 校验参数
	var request GetUsersRequest
	err := ctx.ReadQuery(&request)
	if err != nil {
		responseError(ctx, errors.InvalidRequest)
		return
	}
	// 请求
	filter := model.UserFilter{
		Offset: (request.Page - 1) * request.Size,
		Limit:  request.Size,
	}
	users, err := manager.userManager.Filter(filter)
	if err != nil {
		responseError(ctx, err)
		return
	}
	filter.Offset = 0 // count时Offset必须为0
	filter.Limit = 0
	count := manager.userManager.Count(filter)
	responseSuccess(ctx, "data", iris.Map{
		"users": users,
		"total": count,
	})
}

type AddUserRequest struct {
	Username string `json:"name"`
	Password string `json:"password"`
	Level    int    `json:"level"`
}

// HandlerOfAddUser 新增用户处理函数
func (manager *Manager) HandlerOfAddUser(ctx iris.Context) {
	// 权限检查
	self := manager.getUserClaims(ctx)
	if self.Level < model.UserLevelRoot {
		responseError(ctx, errors.PermissionDeny)
		return
	}
	// 校验参数
	var request AddUserRequest
	err := ctx.ReadJSON(&request)
	if err != nil {
		log.Errorf("read request err: %v", err)
		responseError(ctx, errors.InvalidRequest)
		return
	}
	if len(request.Username) == 0 || len(request.Password) == 0 ||
		request.Level > model.UserLevelRoot || request.Level < model.UserLevelGeneral {
		log.Errorf("invalid request: %+v", request)
		responseError(ctx, errors.InvalidRequest)
		return
	}
	// 请求
	user := model.User{
		Name:   request.Username,
		Passwd: request.Password,
		Level:  request.Level,
	}
	err = manager.userManager.Add(&user)
	if err != nil {
		responseError(ctx, err)
		return
	}
	responseSuccess(ctx, "user", user)
}

type SetUserLevelRequest struct {
	UserID uint `json:"id"`
	Level  int  `json:"level"`
}

// HandlerOfSetUserLevel 设置用户权限等级处理函数
func (manager *Manager) HandlerOfSetUserLevel(ctx iris.Context) {
	// 权限检查
	self := manager.getUserClaims(ctx)
	if self.Level < model.UserLevelRoot {
		responseError(ctx, errors.PermissionDeny)
		return
	}
	// 校验参数
	var request SetUserLevelRequest
	err := ctx.ReadJSON(&request)
	if err != nil {
		responseError(ctx, errors.InvalidRequest)
		return
	}
	if request.Level > model.UserLevelRoot || request.Level < model.UserLevelGeneral {
		responseError(ctx, errors.InvalidRequest)
		return
	}
	if request.UserID == 0 || request.UserID == self.ID {
		responseError(ctx, errors.InvalidRequest)
		return
	}
	// 请求
	err = manager.userManager.SetLevel(request.UserID, request.Level)
	if err != nil {
		responseError(ctx, err)
		return
	}
	responseSuccess(ctx, "", nil)
}

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
	// 保证被冻结用户的权限级别低于调用者
	user, err := manager.userManager.Get(request.UserID)
	if err != nil {
		responseError(ctx, err)
		return
	}
	if user.Level >= self.Level {
		responseError(ctx, errors.PermissionDeny)
		return
	}
	// 冻结该用户的所有实例
	if request.IsFrozen {
		if err = manager.freezeUserInstances(request.UserID); err != nil {
			responseError(ctx, err)
			return
		}
	}
	// 冻结或解冻
	err = manager.userManager.Freeze(request.UserID, request.IsFrozen)
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

// HandlerOfDeleteUser 删除用户处理函数
func (manager *Manager) HandlerOfDeleteUser(ctx iris.Context) {
	// 权限检查
	self := manager.getUserClaims(ctx)
	if self.Level < model.UserLevelRoot {
		responseError(ctx, errors.PermissionDeny)
		return
	}
	// 校验参数
	id := uint(ctx.URLParamUint64("id"))
	if id == 0 || self.ID == id {
		responseError(ctx, errors.InvalidRequest)
		return
	}
	// 冻结用户的所有实例
	if err := manager.freezeUserInstances(id); err != nil {
		responseError(ctx, err)
		return
	}
	// 删除用户
	err := manager.userManager.Delete(id)
	if err != nil {
		responseError(ctx, err)
		return
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
	if len(request.NewPassword) == 0 {
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
