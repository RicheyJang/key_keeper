package safer

import (
	"math"
	"time"
)

type ModelInstance struct {
	ID         uint   `gorm:"primaryKey"`
	Identifier string `gorm:"column:identifier;uniqueIndex"`
	Key        []byte `gorm:"column:key"`
}

func (ins ModelInstance) TableName() string {
	return "t_safer_instances"
}

type ModelKey struct {
	ID         uint   `gorm:"primaryKey;autoIncrement:false"`
	Identifier string `gorm:"primaryKey"`
	Length     uint
	Algorithm  string
	Rotation   uint
	SS         []byte
	CreatedAt  time.Time
}

func (key ModelKey) TableName() string {
	return "t_safer_keys"
}

func (key ModelKey) currentVersion() uint {
	return key.versionAt(time.Now())
}

func (key ModelKey) versionAt(t time.Time) (version uint) {
	if key.Rotation == 0 {
		return 1
	}
	defer func() {
		if version > math.MaxUint32 {
			version = math.MaxUint32
		}
	}()
	passed := t.Sub(key.CreatedAt)
	passVersion := uint(passed/time.Second) / key.Rotation
	return 1 + passVersion
}

func (key ModelKey) nextTimeout() uint {
	return key.nextTimeoutAt(time.Now())
}

func (key ModelKey) nextTimeoutOf(version uint) uint {
	if key.Rotation == 0 {
		return 0
	}
	return uint(key.CreatedAt.Add(time.Duration(version) * time.Duration(key.Rotation) * time.Second).Unix())
}

func (key ModelKey) nextTimeoutAt(t time.Time) uint {
	if key.Rotation == 0 {
		return 0
	}
	version := key.versionAt(t)
	return key.nextTimeoutOf(version)
}
