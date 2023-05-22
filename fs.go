package main

import (
	"fmt"
	"io"
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

type ChunkFileData struct {
	data      [][]byte
	name      string
	extension string
}

const FiveGB = 5 * 1024 * 1024 * 1024

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

func GetChunkFileData(src string) (ChunkFileData, error) {
	file, err := os.Open(src)
	if err != nil {
		return ChunkFileData{}, fmt.Errorf("[readFile][os.Open]: %w", err)
	}

	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return ChunkFileData{}, fmt.Errorf("[readFile][file.Stat]: %w", err)
	}

	fileSize := info.Size()
	numberOfChunks := fileSize / FiveGB
	if fileSize%FiveGB != 0 {
		numberOfChunks++
	}

	chunks := make([][]byte, numberOfChunks)

	for i := int64(0); i < numberOfChunks; i++ {
		buffer := make([]byte, FiveGB)
		bytesRead, err := io.ReadFull(file, buffer)
		if err == io.ErrUnexpectedEOF || err == io.EOF {
			// This will happen on the last chunk if it's smaller than 5GB
			buffer = buffer[:bytesRead]
		} else if err != nil {
			return ChunkFileData{}, fmt.Errorf("[readFile][ io.ReadFull]: %w", err)
		}

		chunks[i] = buffer
	}

	name := filepath.Base(src)
	extension := filepath.Ext(src)
	extension = extension[1:]

	return ChunkFileData{
		data:      chunks,
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
