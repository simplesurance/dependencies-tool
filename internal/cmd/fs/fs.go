package fs

import (
	"fmt"
	"os"
)

type PathType int

const (
	PathTypeUndefined PathType = iota
	PathTypeDir
	PathTypeFile
)

func FileOrDir(path string) (PathType, error) {
	info, err := os.Stat(path)
	if err != nil {
		return -1, err
	}

	if info.IsDir() {
		return PathTypeDir, nil
	}

	if info.Mode().IsRegular() {
		return PathTypeFile, nil
	}

	return -1, fmt.Errorf("path isn't a directory nor a regular file")
}
