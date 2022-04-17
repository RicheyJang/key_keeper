package keeper

import (
	"time"

	"gorm.io/gorm"
)

// KeyRequest 密钥请求
type KeyRequest struct {
	ID      uint `json:"id"`
	Version uint `json:"version"`
}

// KeyInfo 密钥信息
type KeyInfo struct {
	ID        uint   `json:"id"`        // 密钥ID
	Version   uint   `json:"version"`   // 密钥版本
	Key       string `json:"key"`       // 密钥内容（以16进制字符串格式）
	Length    uint   `json:"length"`    // 密钥长度
	Algorithm string `json:"algorithm"` // 加密算法
	Timeout   uint   `json:"timeout"`   // 超时时间戳（需轮替）
}

// KeysFilter 过滤要求
type KeysFilter struct {
	Offset  int  `json:"offset"`
	Limit   int  `json:"limit"`
	Content bool `json:"content"` // keys是否返回密钥内容
}

// DistributeKeyRequest 密钥派发请求
type DistributeKeyRequest struct {
	ID        uint          `json:"id"`
	Length    uint          `json:"length"`       // 密钥长度
	Algorithm string        `json:"algorithm"`    // 加密算法
	Rotation  time.Duration `json:"rotationTime"` // 轮替时长（为0则不轮替）
}

// KeyKeeper 密钥保管器：负责生成密钥、加密保存自己的密钥集、备份密钥等
type KeyKeeper interface {
	GetKeyInfo(request KeyRequest) (KeyInfo, error)
	GetLatestVersionKey(ID uint) (KeyInfo, error)

	FilterKeys(filter KeysFilter) (keys []KeyInfo, total int64, err error)
	DistributeKey(request DistributeKeyRequest) (KeyInfo, error)
	DestroyKey(id uint) error

	Destroy() error // 熔断
}

// Option 生成Keeper时的参数
type Option struct {
	Identifier string
	DB         *gorm.DB
}

// Generator 生成器，用于生成一个Keeper实例
type Generator func(option Option) (KeyKeeper, error)
