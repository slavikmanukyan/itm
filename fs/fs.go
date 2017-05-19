package fs

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/slavikmanukyan/itm/hash"
	"github.com/slavikmanukyan/itm/itmconfig"
	"github.com/slavikmanukyan/itm/utils"
)

// CopyFile copy file
func CopyFile(src, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
// Symlinks are ignored and skipped.
func CopyDir(src string, dst string, config itmconfig.ITMConfig, timestamp string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	// _, err = os.Stat(dst)
	// if err != nil && !os.IsNotExist(err) {
	// 	return
	// }
	// if err == nil {
	// 	return fmt.Errorf("destination already exists")
	// }

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		a1, _ := filepath.Abs(srcPath)
		a2, _ := filepath.Abs(dst)

		if (a1 == a2) || (config.IGNORE[a1]) {
			continue
		}

		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath, config, timestamp)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = CopyFile(srcPath, dstPath)
			if timestamp != "" {
				if config.IS_TERMINAL {
					utils.ClearLine()
				}
				fmt.Print("Copying file: ", srcPath)
				if err != nil {
					return
				}
				hash.SaveFileHash(a1, config, timestamp)
			}
		}
	}

	return
}

func WalkAll(dir string, config itmconfig.ITMConfig) ([]string, error) {
	dir = filepath.Clean(dir)

	si, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !si.IsDir() {
		return nil, fmt.Errorf("source is not a directory")
	}

	var files []string
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(dir, entry.Name())

		a1, _ := filepath.Abs(srcPath)

		if config.IGNORE[a1] {
			continue
		}

		if entry.IsDir() {
			newFiles, err := WalkAll(srcPath, config)
			if err != nil {
				return nil, err
			}
			files = append(files, newFiles...)
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}
			files = append(files, a1)
		}
	}

	return files, nil
}

func ReadFileSlice(file string, n int, size int) []byte {
	slice := make([]byte, size)
	in, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	section := io.NewSectionReader(in, int64(n)*int64(size), int64(size))
	count, _ := section.Read(slice)
	return slice[:count]
}

func WriteFileSlice(file string, index int, size int, data []byte, count int) {
	in, err := os.OpenFile(file, os.O_RDWR, 0644)

	if err == nil {
		defer in.Close()
		offset := int64(index) * int64(size)
		in.WriteAt(data, offset)
	}
}

func RemoveEmptyDirs(dir string, config itmconfig.ITMConfig) {
	var dirs []string

	filepath.Walk(dir, func(file string, info os.FileInfo, err error) error {
		if info.IsDir() {
			dirs = append(dirs, file)
		}
		return nil
	})

	for file := range utils.ReverseChan(dirs) {
		entries, _ := ioutil.ReadDir(file)
		if len(entries) == 0 {
			os.RemoveAll(file)
		}
	}
}
