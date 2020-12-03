package db

import "errors"

var (
	errNotEnoughSkillPoints = errors.New("not enough skill points")
	errWrongStat            = errors.New("wrong stat")
)
