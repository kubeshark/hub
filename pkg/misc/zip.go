package misc

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func ZipIt(dir string) (zipName string, zipPath string, err error) {
	zipName = fmt.Sprintf("kubeshark_%d.zip", time.Now().UnixNano())
	zipPath = BuildDataFilePath("", zipName)
	var file *os.File
	file, err = os.Create(zipPath)
	if err != nil {
		return
	}
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		fmt.Printf("dir: %v\n", dir)
		fmt.Printf("path: %v\n", path[len(dir)+1:])
		f, err := w.Create(path[len(dir)+1:])
		if err != nil {
			return err
		}

		_, err = io.Copy(f, file)
		if err != nil {
			return err
		}

		return nil
	}

	err = filepath.Walk(dir, walker)
	return
}
