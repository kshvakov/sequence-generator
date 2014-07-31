package generator

import (
	"os"
)

func IsDir(name string) bool {

	fileInfo, err := os.Stat(name)

	return err == nil && fileInfo.IsDir()
}

func IsExist(name string) bool {
	return !(IsNotExist(name))
}

func IsNotExist(name string) bool {

	_, err := os.Stat(name)

	return err != nil && os.IsNotExist(err)
}
