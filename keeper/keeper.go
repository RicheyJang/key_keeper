package keeper

// KeyRequest 密钥请求
type KeyRequest struct {
	ID      uint
	Version uint
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

// KeyKeeper 密钥保管器：负责生成密钥、加密保存自己的密钥集、备份密钥等
type KeyKeeper interface {
	GetKeyInfo(request KeyRequest) (KeyInfo, error)
	GetLatestKeyVersion(request KeyRequest) (uint, error)

	Destroy() error // 熔断
}

// Option 生成Keeper时的参数
type Option struct {
	Identifier string
}

// Generator 生成器，用于生成一个Keeper实例
type Generator func(option Option) (KeyKeeper, error)
