package example

import (
	"math"

	"github.com/RicheyJang/key_keeper/keeper"
)

type KeeperEx struct{}

func (k KeeperEx) GetKeyInfo(request keeper.KeyRequest) (keeper.KeyInfo, error) {
	return keeper.KeyInfo{
		ID:        request.ID,
		Version:   request.Version,
		Key:       "98f318d8ed245606332427ea92011ec0",
		Length:    16,
		Algorithm: "aes-cbc",
		Timeout:   math.MaxUint32,
	}, nil
}

func (k KeeperEx) GetLatestVersionKey(ID uint) (keeper.KeyInfo, error) {
	return keeper.KeyInfo{
		ID:        ID,
		Version:   1,
		Key:       "98f318d8ed245606332427ea92011ec0",
		Length:    16,
		Algorithm: "aes-cbc",
		Timeout:   math.MaxUint32,
	}, nil
}

func (k KeeperEx) Destroy() error {
	return nil
}

func NewExampleKeeper(option keeper.Option) (keeper.KeyKeeper, error) {
	return new(KeeperEx), nil
}
