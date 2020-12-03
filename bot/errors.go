package bot

import "errors"

var (
	errNotGameMaster         = errors.New("you are not the game master")
	errCharacterDoesNotExist = errors.New("character doesn't exist")
	errIllegalArgument       = errors.New("illegal argument")
)
