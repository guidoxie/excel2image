package main

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/godruoyi/go-snowflake"
	"github.com/spf13/cast"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

const (
	port = 12128
)

var (
	defaultQuality = 94
	defaultWidth   = 1024
	defaultHeight  = 0
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	snowflake.SetMachineID(0)
	router := gin.New()
	router.Use(gin.Recovery())
	router.POST("api/upload", func(c *gin.Context) {
		param := struct {
			Format  string `form:"format" binding:"oneof=png jpeg jpg"`
			Quality *int   `form:"quality"`
			Width   *int   `form:"width"`
			Height  *int   `form:"height"`
		}{}
		if err := c.ShouldBindQuery(&param); err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		if param.Quality == nil {
			param.Quality = &defaultQuality
		}
		if param.Width == nil {
			param.Width = &defaultWidth
		}
		if param.Height == nil {
			param.Height = &defaultHeight
		}

		f, fh, err := c.Request.FormFile("file")
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		defer f.Close()

		body := &bytes.Buffer{}
		_, err = io.Copy(body, f)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		file, err := excel2image(body.Bytes(), param.Format, *param.Quality, *param.Width, *param.Height)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		// 去掉文件后缀
		fileName := strings.TrimSuffix(fh.Filename, filepath.Ext(fh.Filename))
		// 文件流返回
		c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename=%s`, fmt.Sprintf("%s.%s", url.QueryEscape(fileName), param.Format)))
		if _, err := io.Copy(c.Writer, bytes.NewBuffer(file)); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	})
	if err := router.Run(fmt.Sprintf(":%d", port)); err != nil {
		panic(err)
	}
}

func excel2image(file []byte, format string, quality int, width int, height int) ([]byte, error) {
	var (
		uid   = snowflake.ID()
		xlsx  = path.Join(os.TempDir(), fmt.Sprintf("%d.xlsx", uid))
		html  = path.Join(os.TempDir(), fmt.Sprintf("%d.html", uid))
		image = path.Join(os.TempDir(), fmt.Sprintf("%d.%s", uid, format))
	)
	defer func() {
		for _, f := range []string{xlsx, html, image} {
			_ = os.Remove(f)
		}
	}()
	// 保存文件到临时目录
	if err := WriteFile(xlsx, file, 0666); err != nil {
		return nil, err
	}
	// xlsx => html
	cmd := exec.Command("libreoffice7.6", "--nologo", "--headless", "--convert-to", "html", xlsx, "--outdir", os.TempDir())
	if out, err := cmd.CombinedOutput(); err != nil && !strings.Contains(string(out), "Warning") {
		log.Printf("xlsx => html err: %v out:%s", err, string(out))
		return nil, err
	}
	// html => image
	cmd = exec.Command("wkhtmltoimage", "--quality", cast.ToString(quality), "--width", cast.ToString(width), "--height", cast.ToString(height), "-f", format, html, image)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Printf("html => image err: %v out:%s", err, string(out))
	}
	return ReadFile(image)
}

func WriteFile(name string, data []byte, perm os.FileMode) error {
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}
	return err
}

func ReadFile(name string) ([]byte, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var size int
	if info, err := f.Stat(); err == nil {
		size64 := info.Size()
		if int64(int(size64)) == size64 {
			size = int(size64)
		}
	}
	size++ // one byte for final read at EOF

	// If a file claims a small size, read at least 512 bytes.
	// In particular, files in Linux's /proc claim size 0 but
	// then do not work right if read in small pieces,
	// so an initial read of 1 byte would not work correctly.
	if size < 512 {
		size = 512
	}

	data := make([]byte, 0, size)
	for {
		if len(data) >= cap(data) {
			d := append(data[:cap(data)], 0)
			data = d[:len(data)]
		}
		n, err := f.Read(data[len(data):cap(data)])
		data = data[:len(data)+n]
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return data, err
		}
	}
}
