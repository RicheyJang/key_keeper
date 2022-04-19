package example

import (
	"math"

	"github.com/RicheyJang/key_keeper/keeper"
	"github.com/RicheyJang/key_keeper/utils/errors"
)

type KeeperEx struct{}

var templateError = errors.KeeperNotSupport
var staticKey = keeper.KeyInfo{
	ID:        1,
	Version:   1,
	Key:       "98f318d8ed245606332427ea92011ec0",
	Length:    16,
	Algorithm: "aes-cbc",
	Timeout:   math.MaxUint32,
}

func (k KeeperEx) GetKeyInfo(request keeper.KeyRequest) (keeper.KeyInfo, error) {
	key := staticKey
	key.ID = request.ID
	key.Version = request.Version
	return key, nil
}

func (k KeeperEx) GetLatestVersionKey(ID uint) (keeper.KeyInfo, error) {
	key := staticKey
	key.ID = ID
	return key, nil
}

func (k KeeperEx) FilterKeys(filter keeper.KeysFilter) (keys []keeper.KeyInfo, total int64, err error) {
	if filter.Offset == 0 && filter.Limit > 0 {
		keys = []keeper.KeyInfo{staticKey}
	}
	return keys, 1, nil
}

func (k KeeperEx) DistributeKey(request keeper.DistributeKeyRequest) (keeper.KeyInfo, error) {
	return staticKey, templateError
}

func (k KeeperEx) FreezeKey(id uint, beFrozen bool) error {
	return templateError
}

func (k KeeperEx) DestroyKey(id uint) error {
	return templateError
}

func (k KeeperEx) Destroy() error {
	return nil
}

func NewExampleKeeper(option keeper.Option) (keeper.KeyKeeper, error) {
	return new(KeeperEx), nil
}
