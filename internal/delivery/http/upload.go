package http

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Uploader struct {
	dir string
}

func NewUploader(dir string) (*Uploader, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &Uploader{dir: dir}, nil
}

func (u *Uploader) Save(files []*multipart.FileHeader) ([]string, error) {
	paths := make([]string, 0, len(files))
	for _, file := range files {
		if file == nil {
			continue
		}
		name := safeFilename(file.Filename)
		stored := fmt.Sprintf("%d-%s", time.Now().UnixNano(), name)
		target := filepath.Join(u.dir, stored)
		if err := saveUploadedFile(file, target); err != nil {
			return nil, err
		}
		paths = append(paths, filepath.ToSlash(filepath.Join(u.dir, stored)))
	}
	return paths, nil
}

func saveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = out.ReadFrom(src)
	return err
}

func safeFilename(name string) string {
	name = filepath.Base(name)
	name = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '.' || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, name)
	if name == "" || name == "." {
		return "upload.bin"
	}
	return name
}

func multipartFiles(c *gin.Context, key string) []*multipart.FileHeader {
	form, err := c.MultipartForm()
	if err != nil || form == nil {
		return nil
	}
	return form.File[key]
}
