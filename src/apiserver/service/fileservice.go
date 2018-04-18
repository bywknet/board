package service

import (
	"archive/zip"
	"fmt"
	"git/inspursoft/board/src/common/model"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	metaFile = "META.cfg"
)

func ListUploadFiles(directory string) ([]model.FileInfo, error) {
	uploads := []model.FileInfo{}
	filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if info != nil && !info.IsDir() {
			uploads = append(uploads, model.FileInfo{
				Path:     filepath.Dir(path),
				FileName: info.Name(),
				Size:     info.Size(),
			})
		}
		return err
	})
	return uploads, nil
}

func RemoveUploadFile(file model.FileInfo) error {
	return os.Remove(filepath.Join(file.Path, file.FileName))
}

func CreateMetaConfiguration(configurations map[string]string, targetPath string) error {
	if configurations == nil {
		return fmt.Errorf("configuration for generating base directory is nil")
	}
	f, err := os.OpenFile(filepath.Join(targetPath, metaFile), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create META.cfg file: %+v", err)
	}
	defer f.Close()
	f.WriteString("[para]\n")
	for key, value := range configurations {
		fmt.Fprintf(f, "%s=%s\n", key, value)
	}
	return nil
}

func CopyFile(sourcePath, targetPath string) error {
	from, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer from.Close()

	to, err := os.OpenFile(targetPath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		return err
	}
	return nil
}

func CreateFile(fileName, message, targetPath string) error {
	f, err := os.OpenFile(filepath.Join(targetPath, fileName), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create %s file: %+v", fileName, err)
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "%s\n", message)
	return err
}

func ZipFiles(zipFileName, dirName string) error {
	zipFile, err := os.OpenFile(zipFileName, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer zipFile.Close()
	w := zip.NewWriter(zipFile)
	err = filepath.Walk(dirName, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			f, err := w.Create(info.Name())
			if err != nil {
				return err
			}
			data, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			_, err = f.Write(data)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return w.Close()
}
