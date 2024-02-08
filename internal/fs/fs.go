package fs

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/simplesurance/dependencies-tool/v3/internal/datastructs"
)

// Find searches in rootdir and its sub-directories for files that have the
// suffix matchSuffix.
// If matchSuffix contains more then 1 path element it must be separated by
// filepath.Separator.
// It returns the paths to all found files.
// Directories that are named as an element in ignoredDirs are not searched.
// Symlinks are not followed.
func Find(rootdir string, matchSuffix string, ignoredDirs []string) ([]string, error) {
	var res []string

	ignored := datastructs.SliceToSet(ignoredDirs)

	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if _, skip := ignored[filepath.Base(path)]; skip {
				return fs.SkipDir
			}
			return nil
		}

		if strings.HasSuffix(path, matchSuffix) {
			res = append(res, path)
		}

		return nil
	}

	err := filepath.WalkDir(rootdir, walkFn)
	if err != nil {
		return nil, err
	}

	return res, nil
}
