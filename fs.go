package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	lop "github.com/samber/lo/parallel"
)

type FileData struct {
	data      []byte
	name      string
	extension string
}

func GetFileData(src string) (FileData, error) {
	data, err := os.ReadFile(src)
	if err != nil {
		return FileData{}, fmt.Errorf("[readFile][ioutil.ReadFile]: %w", err)
	}

	name := filepath.Base(src)
	extension := filepath.Ext(src)
	extension = extension[1:]

	return FileData{
		data:      data,
		name:      name,
		extension: extension,
	}, nil
}

func GetDirFiles(src string) ([]string, error) {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return nil, fmt.Errorf("[os.Stat]: %w", err)
	}

	dir := filepath.Dir(src)

	// otherwise nested forlder will be dropped from path
	if srcInfo.IsDir() {
		dir = src
	}

	files, err := os.ReadDir(src)
	if err != nil {
		return nil, fmt.Errorf("[os.ReadDir]: %w", err)
	}

	var paths []string

	lop.ForEach(files, func(file fs.DirEntry, _ int) {
		info, err := file.Info()
		if err != nil {
			log.Printf("[file.Info]: %s", err)
		} else if info.IsDir() {
			// recursively get nested files
			fullNestedPath := fmt.Sprintf("%s%s", dir, info.Name())
			nestedFiles, err := GetDirFiles(fullNestedPath)
			if err != nil {
				log.Printf("[GetDirFiles](nestedFiles): %s", err)
			} else {
				paths = append(paths, nestedFiles...)
			}
		} else if info.Name() == ".DS_Store" {
			return
		} else {
			fullPath := fmt.Sprintf("%s/%s", dir, info.Name())
			paths = append(paths, fullPath)
		}
	})

	return paths, nil
}
