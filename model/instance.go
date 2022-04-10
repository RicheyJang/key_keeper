package model

import (
	"strings"
	"time"

	"github.com/RicheyJang/key_keeper/utils/errors"
)

type Instance struct {
	ID         uint   `gorm:"primaryKey"`
	Identifier string `gorm:"column:identifier;uniqueIndex"`
	Keeper     string `gorm:"column:keeper"`
	Users      string `gorm:"column:users"`
	DSafeLevel int    `gorm:"column:d_safe_level"`
	IPs        string `gorm:"column:ips"`
	CreatedAt  time.Time
}

func (ins Instance) TableName() string {
	return "t_manager_instances"
}

const InstanceUserDelimiter = ","

func (ins Instance) GetUsers() []string {
	return strings.Split(ins.Users, InstanceUserDelimiter)
}

func (ins Instance) HasUser(user string) bool {
	users := ins.GetUsers()
	for _, old := range users {
		if old == user {
			return true
		}
	}
	return false
}

func (ins *Instance) AddUser(user string) error {
	if strings.Contains(user, InstanceUserDelimiter) {
		return errors.New(errors.CodeRequest, "username cannot contain commas")
	}
	if strings.HasSuffix(ins.Users, InstanceUserDelimiter) {
		ins.Users += user
	} else {
		ins.Users += InstanceUserDelimiter + user
	}
	return nil
}

func (ins *Instance) DeleteUser(user string) {
	users := ins.GetUsers()
	var newUsers []string
	for _, old := range users {
		if old != user {
			newUsers = append(newUsers, old)
		}
	}
	ins.Users = strings.Join(newUsers, InstanceUserDelimiter)
}
