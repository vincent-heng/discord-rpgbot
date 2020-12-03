package util

import (
	"bufio"
	"io/ioutil"
	"os"
	"strconv"
)

func fileExists(fileName string) bool {
	_, err := os.Stat(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func GetChannelID() (string, error) {
	return readFile("current_channel.txt")
}

func readFile(fileName string) (string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return "", nil
	}
	scanner := bufio.NewScanner(file)

	success := scanner.Scan()
	if !success {
		// False on error or EOF. Check error
		err = scanner.Err()
		if err == nil {
			return "", nil
		}
		return "", err
	}

	return scanner.Text(), nil
}

func DiscordIDToText(userID uint) string {
	return "<@" + strconv.FormatUint(uint64(userID), 10) + ">"
}

func SetAdventureChannel(channelID string) error {
	fileName := "current_channel.txt"
	if !fileExists(fileName) {
		file, err := os.Create(fileName)
		if err != nil {
			return err
		}
		file.Close()
	} else {
		err := os.Truncate(fileName, 0)
		if err != nil {
			return err
		}
	}

	err := ioutil.WriteFile(fileName, []byte(channelID), 0666)
	if err != nil {
		return err
	}

	return nil
}
