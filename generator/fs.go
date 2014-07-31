package generator

import (
	"os"
)

func isDir(name string) bool {

	fileInfo, err := os.Stat(name)

	return err == nil && fileInfo.IsDir()
}

func isNotExist(name string) bool {

	_, err := os.Stat(name)

	return err != nil && os.IsNotExist(err)
}
