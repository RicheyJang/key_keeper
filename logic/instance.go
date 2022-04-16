package logic

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"

	"github.com/RicheyJang/key_keeper/keeper"
	"github.com/RicheyJang/key_keeper/model"
	"github.com/RicheyJang/key_keeper/utils/errors"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

// InstanceInfo 实例信息
type InstanceInfo struct {
	model.Instance
	kp keeper.KeyKeeper
}

const DefaultInstanceIdentifier = "default"

type GetInstancesRequest struct {
	Page int `url:"page"`
	Size int `url:"size"`
}

// HandlerOfGetInstances 批量获取指定实例处理函数
func (manager *Manager) HandlerOfGetInstances(ctx iris.Context) {
	// 校验参数
	var request GetInstancesRequest
	err := ctx.ReadQuery(&request)
	if err != nil {
		responseError(ctx, errors.InvalidRequest)
		return
	}
	// 预处理请求
	offset, limit := (request.Page-1)*request.Size, request.Size
	if offset < 0 {
		offset = 0
	}
	var instances []model.Instance
	var count int64
	// 分情况处理
	self := manager.getUserClaims(ctx)
	if self.Level >= model.UserLevelRoot { // Root用户：返回所有实例
		if err = manager.db.Model(&model.Instance{}).Count(&count).Error; err != nil {
			responseError(ctx, err)
			return
		}
		session := manager.db.Order("id")
		if offset > 0 {
			session = session.Offset(offset)
		}
		if limit > 0 {
			session = session.Limit(limit)
		}
		if err = session.Find(&instances).Error; err != nil {
			responseError(ctx, err)
			return
		}
	} else { // 其它用户：返回自己管理的实例
		instances, err = manager.getInstancesByUser(self.ID)
		if err != nil {
			responseError(ctx, err)
			return
		}
		count = int64(len(instances))
		if limit > 0 && count > 0 { // 分页
			if offset+limit > int(count) {
				instances = instances[offset:]
			} else {
				instances = instances[offset : offset+limit]
			}
		}
	}
	// 回包
	if instances == nil {
		instances = make([]model.Instance, 0)
	}
	responseSuccess(ctx, "data", iris.Map{
		"instances": instances,
		"total":     count,
	})
}

type AddInstanceRequest struct {
	Identifier string `json:"identifier"`
	DSafeLevel int    `json:"level"`
	IPs        string `json:"ips"`
	Keeper     string `json:"keeper"`
}

var instanceIdentifierRegexp = regexp.MustCompile(`^[\w.+-]+$`)

// HandlerOfAddInstance 添加实例处理函数
func (manager *Manager) HandlerOfAddInstance(ctx iris.Context) {
	// 校验参数
	var request AddInstanceRequest
	if err := ctx.ReadJSON(&request); err != nil {
		responseError(ctx, errors.InvalidRequest)
		return
	}
	if len(request.Identifier) == 0 || len(request.Keeper) == 0 {
		responseError(ctx, errors.InvalidRequest)
		return
	}
	if !instanceIdentifierRegexp.MatchString(request.Identifier) {
		responseError(ctx, errors.InvalidRequest)
		return
	}
	if _, ok := manager.getInstance(request.Identifier); ok == true {
		responseError(ctx, errors.InstanceExist)
		return
	}
	// 创建实例
	self := manager.getUserClaims(ctx)
	instance := model.Instance{
		Identifier: request.Identifier,
		IsFrozen:   false,
		Keeper:     request.Keeper,
		DSafeLevel: request.DSafeLevel,
		IPs:        request.IPs,
	}
	if err := instance.AddUser(strconv.FormatUint(uint64(self.ID), 10)); err != nil {
		responseError(ctx, err)
		return
	}
	if _, err := manager.createInstance(instance); err != nil {
		responseError(ctx, err)
		return
	}
	responseSuccess(ctx, "instance", instance)
}

type FreezeInstanceRequest struct {
	Identifier string `json:"identifier"`
	IsFrozen   bool   `json:"isFrozen"`
}

// HandlerOfFreezeInstance 冻结指定实例处理函数
func (manager *Manager) HandlerOfFreezeInstance(ctx iris.Context) {
	// 校验参数
	var request FreezeInstanceRequest
	err := ctx.ReadJSON(&request)
	if err != nil {
		responseError(ctx, errors.InvalidRequest)
		return
	}
	if len(request.Identifier) == 0 || request.Identifier == DefaultInstanceIdentifier {
		responseError(ctx, errors.InvalidRequest)
		return
	}
	// 获取实例
	_, err = manager.getInstanceAndCheckUser(request.Identifier, ctx)
	if err != nil {
		responseError(ctx, err)
		return
	}
	// 冻结实例
	if err = manager.freezeInstance(request.Identifier, request.IsFrozen); err != nil {
		responseError(ctx, err)
		return
	}
	// 回包
	responseSuccess(ctx, "", nil)
}

// HandlerOfDestroyInstance 销毁实例处理函数
func (manager *Manager) HandlerOfDestroyInstance(ctx iris.Context) {
	// 校验参数
	identifier := ctx.URLParam("identifier")
	if len(identifier) == 0 || identifier == DefaultInstanceIdentifier {
		responseError(ctx, errors.InvalidRequest)
		return
	}
	// 获取实例
	instance, err := manager.getInstanceAndCheckUser(identifier, ctx)
	if err != nil {
		responseError(ctx, err)
		return
	}
	// 销毁实例
	err = manager.db.Transaction(func(tx *gorm.DB) (txErr error) {
		defer func() {
			if r := recover(); r != nil {
				txErr = fmt.Errorf("panic: %v", r)
			}
		}()
		if txErr = tx.Where("identifier = ?", identifier).Delete(&model.Instance{}).Error; txErr != nil {
			return
		}
		if txErr = instance.kp.Destroy(); txErr != nil {
			return
		}
		return
	})
	if err != nil {
		responseError(ctx, err)
		return
	}
	manager.instanceMap.Delete(identifier)
	// 回包
	responseSuccess(ctx, "", nil)
}

// 初始化所有实例
func (manager *Manager) initAllInstances() error {
	// 初始化数据库
	if err := manager.db.AutoMigrate(&model.Instance{}); err != nil {
		return err
	}
	// 读取出所有实例
	var instances []model.Instance
	if err := manager.db.Find(&instances).Error; err != nil {
		return err
	}
	sort.Slice(instances, func(i, j int) bool { // 令默认实例位于首位
		if instances[i].Identifier == DefaultInstanceIdentifier {
			return true
		} else if instances[j].Identifier == DefaultInstanceIdentifier {
			return false
		}
		return instances[i].ID < instances[j].ID
	})
	// 保证默认实例
	if len(instances) == 0 || instances[0].Identifier != DefaultInstanceIdentifier {
		if _, err := manager.createInstance(model.Instance{
			Identifier: DefaultInstanceIdentifier,
			Keeper:     manager.defaultKName,
			Users:      "root",
		}); err != nil {
			return err
		}
	}
	// 初始化所有实例
	for _, instance := range instances {
		err := manager.initInstance(instance)
		if err != nil {
			return err
		}
	}
	return nil
}

// 初始化单个实例
func (manager *Manager) initInstance(instance model.Instance) error {
	// 获取generator
	generatorValue, ok := manager.generatorMap.Load(instance.Keeper)
	if !ok || generatorValue == nil {
		return errors.InvalidKeeper
	}
	generator := generatorValue.(keeper.Generator)
	// 初始化实例
	kp, err := generator(keeper.Option{
		Identifier: instance.Identifier,
		DB:         manager.db,
	})
	if err != nil {
		return err
	}
	// 存储实例信息
	manager.instanceMap.Store(instance.Identifier, &InstanceInfo{
		Instance: instance,
		kp:       kp,
	})
	return nil
}

// 创建实例
func (manager *Manager) createInstance(instance model.Instance) (InstanceInfo, error) {
	// 获取generator
	generatorValue, ok := manager.generatorMap.Load(instance.Keeper)
	if !ok || generatorValue == nil {
		return InstanceInfo{}, errors.InvalidKeeper
	}
	generator := generatorValue.(keeper.Generator)
	var kp keeper.KeyKeeper
	// 创建实例
	err := manager.db.Transaction(func(tx *gorm.DB) (txErr error) {
		defer func() {
			if r := recover(); r != nil {
				txErr = fmt.Errorf("panic: %v", r)
			}
		}()
		// 数据库创建
		if txErr = tx.Create(&instance).Error; txErr != nil {
			return
		}
		// 生成keeper
		kp, txErr = generator(keeper.Option{
			Identifier: instance.Identifier,
			DB:         manager.db,
		})
		return
	})
	if err != nil {
		return InstanceInfo{}, err
	}
	// 存储实例信息
	info := InstanceInfo{
		Instance: instance,
		kp:       kp,
	}
	manager.instanceMap.Store(instance.Identifier, &info)
	return info, nil
}

// 冻结实例
func (manager *Manager) freezeInstance(identifier string, isFrozen bool) error {
	info, ok := manager.getInstance(identifier)
	if !ok {
		return errors.NoSuchInstance
	}
	if err := manager.db.Model(&model.Instance{}).
		Where("identifier = ?", identifier).Update("is_frozen", isFrozen).Error; err != nil {
		return err
	}
	info.IsFrozen = isFrozen
	return nil
}

// 冻结指定用户直接管理的所有实例
func (manager *Manager) freezeUserInstances(id uint) error {
	// 获取用户直接管理的实例列表
	instances, err := manager.getInstancesByUser(id)
	if err != nil {
		return err
	}
	// 冻结所有实例
	for _, instance := range instances {
		if instance.Identifier == DefaultInstanceIdentifier { // 除了默认实例
			continue
		}
		if err = manager.freezeInstance(instance.Identifier, true); err != nil {
			return err
		}
	}
	return nil
}

// 查询用户直接管理的实例列表（不区分root和普通用户）
func (manager *Manager) getInstancesByUser(id uint) ([]model.Instance, error) {
	idStr := strconv.FormatUint(uint64(id), 10)
	var preSelect []model.Instance
	if err := manager.db.Where("users LIKE ?", "%"+idStr+"%").Order("id").Find(&preSelect).Error; err != nil {
		return nil, err
	}
	var instances []model.Instance
	for _, pre := range preSelect {
		if pre.HasUser(idStr) {
			instances = append(instances, pre)
		}
	}
	return instances, nil
}

// 获取特定实例
func (manager *Manager) getInstance(identifier string) (*InstanceInfo, bool) {
	if identifier == "" {
		return nil, false
	}
	infoValue, ok := manager.instanceMap.Load(identifier)
	if !ok || infoValue == nil {
		return nil, false
	}
	info, ok := infoValue.(*InstanceInfo)
	return info, ok
}

// 获取特定实例并检查用户对该实例是否有权限
func (manager *Manager) getInstanceAndCheckUser(identifier string, ctx iris.Context) (*InstanceInfo, error) {
	info, ok := manager.getInstance(identifier)
	if !ok {
		return nil, errors.NoSuchInstance
	}
	// 检查用户权限
	user := manager.getUserClaims(ctx)
	if user.Level >= model.UserLevelRoot || info.HasUser(strconv.FormatUint(uint64(user.ID), 10)) {
		return info, nil
	} else {
		return nil, errors.PermissionDeny
	}
}
