package safer

import (
	"encoding/hex"
	"sync"
	"time"

	"github.com/RicheyJang/key_keeper/keeper"
	"github.com/RicheyJang/key_keeper/utils/errors"
	"gorm.io/gorm"
)

var migrateOnce sync.Once

func GetSafer(option keeper.Option) (keeper.KeyKeeper, error) {
	// 参数校验
	if option.DB == nil || len(option.Identifier) == 0 {
		return nil, errors.InvalidRequest
	}
	// 初始化
	var err error
	migrateOnce.Do(func() {
		err = option.DB.AutoMigrate(&ModelInstance{}, &ModelKey{})
	})
	if err != nil {
		return nil, err
	}
	// 从数据库中获取mainKey 或 创建实例记录
	var instance ModelInstance
	result := option.DB.Where("identifier = ?", option.Identifier).Find(&instance)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 { // 暂无该实例，创建
		ss, err := getNewSS(mainKeyLength)
		if err != nil {
			return nil, err
		}
		instance = ModelInstance{
			Identifier: option.Identifier,
			Key:        ss,
		}
		if err = option.DB.Create(&instance).Error; err != nil {
			return nil, err
		}
	}
	return &KeeperSF{
		identifier: option.Identifier,
		db:         option.DB,
		mainKey:    instance.Key,
	}, nil
}

type KeeperSF struct {
	identifier string
	db         *gorm.DB
	mainKey    []byte
}

func (sf *KeeperSF) GetKeyInfo(request keeper.KeyRequest) (keeper.KeyInfo, error) {
	//TODO implement me
	return keeper.KeyInfo{}, errors.KeeperNotSupport
}

func (sf *KeeperSF) GetLatestVersionKey(id uint) (keeper.KeyInfo, error) {
	//TODO implement me
	return keeper.KeyInfo{}, errors.KeeperNotSupport
}

func (sf *KeeperSF) FilterKeys(filter keeper.KeysFilter) (keys []keeper.KeyInfo, total int64, err error) {
	var models []ModelKey
	if err = sf.setupKeysFilter(filter).Find(&models).Error; err != nil {
		return
	}
	now := time.Now()
	for _, model := range models {
		key := keeper.KeyInfo{
			ID:        model.ID,
			Version:   model.versionAt(now),
			Length:    model.Length,
			Algorithm: model.Algorithm,
			Timeout:   model.nextTimeoutAt(now),
		}
		if filter.Content {
			var content []byte
			content, err = getKeyContent(key.Length, key.ID, key.Version, sf.mainKey, model.SS)
			if err != nil {
				return
			}
			key.Key = hex.EncodeToString(content)
		}
		keys = append(keys, key)
	}
	filter.Offset = 0
	filter.Limit = 0
	if err = sf.setupKeysFilter(filter).Count(&total).Error; err != nil {
		return
	}
	return
}

func (sf *KeeperSF) DistributeKey(request keeper.DistributeKeyRequest) (keeper.KeyInfo, error) {
	// 构建新密钥
	ss, err := getNewSS(ssLength)
	if err != nil {
		return keeper.KeyInfo{}, err
	}
	key := ModelKey{
		ID:         request.ID,
		Identifier: sf.identifier,
		Length:     request.Length,
		Algorithm:  request.Algorithm,
		Rotation:   request.Rotation,
		SS:         ss,
	}
	// 保存至数据库
	if err = sf.db.Create(&key).Error; err != nil {
		return keeper.KeyInfo{}, err
	}
	// 获取密钥信息
	content, err := getKeyContent(key.Length, key.ID, 1, sf.mainKey, key.SS)
	if err != nil {
		return keeper.KeyInfo{}, err
	}
	return keeper.KeyInfo{
		ID:        key.ID,
		Version:   1,
		Key:       hex.EncodeToString(content),
		Length:    key.Length,
		Algorithm: key.Algorithm,
		Timeout:   key.nextTimeoutOf(1),
	}, nil
}

func (sf *KeeperSF) DestroyKey(id uint) error {
	return sf.db.Where("identifier = ?", sf.identifier).Where("id = ?", id).Delete(&ModelKey{}).Error
}

func (sf *KeeperSF) Destroy() error {
	txErr := sf.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("identifier = ?", sf.identifier).Delete(&ModelInstance{}).Error; err != nil {
			return err
		}
		if err := tx.Where("identifier = ?", sf.identifier).Delete(&ModelKey{}).Error; err != nil {
			return err
		}
		return nil
	})
	if txErr != nil {
		return txErr
	}
	return nil
}

func (sf *KeeperSF) setupKeysFilter(filter keeper.KeysFilter) *gorm.DB {
	session := sf.db.Model(&ModelKey{}).Where("identifier = ?", sf.identifier).Order("id")
	if filter.Offset > 0 {
		session = session.Offset(filter.Offset)
	}
	if filter.Limit > 0 {
		session = session.Limit(filter.Limit)
	}
	return session
}
