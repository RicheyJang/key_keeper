package logic

import (
	"sort"

	"github.com/RicheyJang/key_keeper/utils/errors"

	"github.com/RicheyJang/key_keeper/keeper"
	"github.com/RicheyJang/key_keeper/model"
)

// InstanceInfo 实例信息
type InstanceInfo struct {
	model.Instance
	kp keeper.KeyKeeper
}

const DefaultInstanceIdentifier = "default"

// LoadAllInstances 初始化所有实例
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
	if !ok {
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
	info := InstanceInfo{
		Instance: instance,
		kp:       kp,
	}
	manager.instanceMap.Store(instance.Identifier, info)
	return nil
}

func (manager *Manager) createInstance(instance model.Instance) (InstanceInfo, error) {
	// 获取generator
	generatorValue, ok := manager.generatorMap.Load(instance.Keeper)
	if !ok {
		return InstanceInfo{}, errors.InvalidKeeper
	}
	generator := generatorValue.(keeper.Generator)
	// 创建实例
	if err := manager.db.Create(&instance).Error; err != nil {
		return InstanceInfo{}, err
	}
	kp, err := generator(keeper.Option{
		Identifier: instance.Identifier,
		DB:         manager.db,
	})
	if err != nil {
		return InstanceInfo{}, err
	}
	// 存储实例信息
	info := InstanceInfo{
		Instance: instance,
		kp:       kp,
	}
	manager.instanceMap.Store(instance.Identifier, info)
	return info, nil
}
