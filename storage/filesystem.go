package storage

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"path"
)

type FileSystemBackend struct {
	BaseStorageBackend

	baseDir string
}

func NewFileSystemBackend(basDir string) *FileSystemBackend {
	if basDir == "" {
		basDir = path.Join(os.TempDir(), "pkgstore")
	}
	if _, err := os.Stat(basDir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(basDir, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	return &FileSystemBackend{
		baseDir: basDir,
	}
}

func (s *FileSystemBackend) WriteFile(key string, fileMeta interface{}, r io.Reader) error {
	fileDir, _ := path.Split(key)
	if fileDir != "" {
		err := os.MkdirAll(path.Join(s.baseDir, fileDir), os.ModePerm)
		if err != nil {
			return err
		}
	}
	f, err := os.OpenFile(path.Join(s.baseDir, key), os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Println(err)
		}
	}(f)
	_, err = io.Copy(f, r)
	if err != nil {
		return err
	}
	if fileMeta == nil {
		return nil
	}

	metaBytes, err := json.Marshal(fileMeta)
	if err != nil {
		return err
	}
	err = os.WriteFile(path.Join(s.baseDir, key+".meta.json"), metaBytes, os.ModePerm)
	return err
}

func (s *FileSystemBackend) GetFile(key string) (io.ReadCloser, error) {
	f, err := os.Open(path.Join(s.baseDir, key))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return f, nil
}

func (s *FileSystemBackend) GetMetadata(key string, value interface{}) error {
	metaBytes, err := os.ReadFile(path.Join(s.baseDir, key+".meta.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	return json.Unmarshal(metaBytes, value)
}

func (s *FileSystemBackend) CopyFile(fromKey, toKey string) error {
	fileDir, _ := path.Split(toKey)
	if fileDir != "" {
		err := os.MkdirAll(path.Join(s.baseDir, fileDir), os.ModePerm)
		if err != nil {
			return err
		}
	}

	fromFile, err := os.Open(path.Join(s.baseDir, fromKey))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}
	defer func(fromFile *os.File) {
		err := fromFile.Close()
		if err != nil {
			log.Println(err)
		}
	}(fromFile)

	toFile, err := os.OpenFile(path.Join(s.baseDir, toKey), os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer func(toFile *os.File) {
		err := toFile.Close()
		if err != nil {
			log.Println(err)
		}
	}(toFile)

	_, err = io.Copy(toFile, fromFile)
	return err
}

func (s *FileSystemBackend) DeleteFile(key string) error {
	err := os.Remove(path.Join(s.baseDir, key))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	err = os.Remove(path.Join(s.baseDir, key+".meta.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
	}
	return err
}
