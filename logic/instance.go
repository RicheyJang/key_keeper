package logic

import "github.com/RicheyJang/key_keeper/keeper"

// InstanceInfo 实例信息
type InstanceInfo struct {
	kp         keeper.KeyKeeper
	Identifier string
	UsersID    []string
	IsFrozen   bool
	IPs        []string
}

const DefaultInstanceID = "default"

func LoadAllInstances(kgs []KeeperGeneratorPair) ([]InstanceInfo, error) {
	// TODO 获取所有实例，其中[0]为默认实例，Identifier=DefaultInstanceID
	return nil, nil
}
