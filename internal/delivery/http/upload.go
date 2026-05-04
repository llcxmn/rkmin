package http

import (
	"fmt"
	"mime/multipart"
	nethttp "net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Uploader struct {
	dir string
}

const maxUploadSize = 5 << 20

var allowedImageExt = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
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
		if err := validateUpload(file); err != nil {
			return nil, err
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

func validateUpload(file *multipart.FileHeader) error {
	if file.Size > maxUploadSize {
		return fmt.Errorf("file %s exceeds maximum size 5MB", file.Filename)
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedImageExt[ext] {
		return fmt.Errorf("file %s must be jpg, jpeg, png, or webp", file.Filename)
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	buf := make([]byte, 512)
	n, err := src.Read(buf)
	if err != nil && n == 0 {
		return err
	}
	contentType := nethttp.DetectContentType(buf[:n])
	if !strings.HasPrefix(contentType, "image/") {
		return fmt.Errorf("file %s must be an image", file.Filename)
	}
	return nil
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
