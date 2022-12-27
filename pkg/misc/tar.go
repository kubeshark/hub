package misc

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func TarIt(dir string) (zipName string, zipPath string, err error) {
	zipName = fmt.Sprintf("kubeshark_%d.tar.gz", time.Now().Unix())
	zipPath = BuildDataFilePath("", zipName)
	var file *os.File
	file, err = os.Create(zipPath)
	if err != nil {
		return
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

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

		stat, err := file.Stat()
		if err != nil {
			return err
		}

		header := &tar.Header{
			Name:    path[len(dir)+1:],
			Size:    stat.Size(),
			Mode:    int64(stat.Mode()),
			ModTime: stat.ModTime(),
		}

		err = tarWriter.WriteHeader(header)
		if err != nil {
			return err
		}

		_, err = io.Copy(tarWriter, file)
		if err != nil {
			return err
		}

		return nil
	}

	err = filepath.Walk(dir, walker)
	return
}
