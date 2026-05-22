package handler

import (
	"image/jpeg"
	"image/png"
	"io"
	"myoss/config"
	"myoss/internal/db"
	"myoss/internal/model"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

// 根据后缀判断是否可压缩
func canCompress(ext string) bool {
	ext = strings.ToLower(ext)
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png"
}

// 创建文件 URL
func buildURL(c *echo.Context, filename string) string {
	if config.Config.App.FileBaseURL != "" {
		return config.Config.App.FileBaseURL + "/" + filename
	} else {
		return c.Scheme() + "://" + c.Request().Host + "/i/" + filename
	}
}

// 获取文件保存路径
func GetUploadDir() string {
	dir := config.Config.App.UploadDir
	if dir == "" {
		dir = "./uploads"
	}

	// 转绝对路径（兼容相对路径）
	abs, err := filepath.Abs(dir)
	if err != nil {
		return dir
	}
	return abs
}

// 压缩图片
func compressImage(src, dst string, quality int) error {
	img, err := imaging.Open(src, imaging.AutoOrientation(true))
	if err != nil {
		return err
	}

	if img.Bounds().Dx() > 1600 {
		img = imaging.Resize(img, 1600, 0, imaging.Lanczos)
	}

	tmp := dst + ".tmp"

	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	ext := strings.ToLower(filepath.Ext(src))
	switch ext {
	case ".jpg", ".jpeg":
		err = jpeg.Encode(f, img, &jpeg.Options{Quality: quality})
	case ".png":
		err = png.Encode(f, img)
	}
	f.Close()

	if err != nil {
		_ = os.Remove(tmp)
		return err
	}

	return os.Rename(tmp, dst)
}

// Web 静态文件服务
func ServeFile(c *echo.Context) error {
	filename := c.Param("filename")

	// 分桶目录
	ext := filepath.Ext(filename)
	uuid := strings.TrimSuffix(filename, ext)
	if len(uuid) != 32 {
		return c.JSON(404, map[string]string{"msg": "文件不存在"})
	}

	dir := filepath.Join(GetUploadDir(), uuid[:2], uuid[2:4])
	path := filepath.Join(dir, filename)

	// 数据库校验
	var f model.File
	err := db.DB.Get(&f, "SELECT is_private FROM files WHERE uuid=?", uuid[:32])
	if err != nil {
		return c.JSON(404, map[string]string{"msg": "文件不存在"})
	}

	// 判断是否私有：如果是私有文件，则需要提供正确的 token
	if f.IsPrivate == 1 {
		token := c.QueryParam("token")
		if token != config.Config.Security.Token {
			return c.JSON(403, map[string]string{"msg": "无访问权限"})
		}
	}

	// 读取并输出文件
	fi, err := os.Stat(path)
	if os.IsNotExist(err) || fi.IsDir() {
		return c.JSON(404, map[string]string{"msg": "文件不存在"})
	}

	ff, err := os.Open(path)
	if err != nil {
		return c.JSON(404, map[string]string{"msg": "文件不存在"})
	}

	http.ServeContent(c.Response(), c.Request(), fi.Name(), fi.ModTime(), ff)
	return nil
}

// 上传图片文件
func Upload(c *echo.Context) error {
	isCompress, _ := strconv.Atoi(c.FormValue("compress")) // 是否压缩
	quality, _ := strconv.Atoi(c.FormValue("quality"))     // 压缩质量
	isPrivate, _ := strconv.Atoi(c.FormValue("private"))   // 是否私有

	if quality < 10 || quality > 100 {
		quality = 75
	}
	if isPrivate != 0 && isPrivate != 1 {
		isPrivate = 0
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(400, map[string]string{"msg": "请上传文件"})
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	uid := strings.ReplaceAll(uuid.NewString(), "-", "")

	// 文件分桶存储
	dir := filepath.Join(GetUploadDir(), uid[:2], uid[2:4])
	if err := os.MkdirAll(dir, 0755); err != nil {
		return c.JSON(500, map[string]string{"msg": "目录创建失败"})
	}

	filename := uid + ext
	savePath := filepath.Join(dir, filename)

	// 保存原文件
	src, err := file.Open()
	if err != nil {
		return c.JSON(500, map[string]string{"msg": "文件读取失败"})
	}
	defer src.Close()

	dst, err := os.Create(savePath)
	if err != nil {
		return c.JSON(500, map[string]string{"msg": "文件创建失败"})
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return c.JSON(500, map[string]string{"msg": "文件保存失败"})
	}

	// 压缩图片，覆盖原文件
	if isCompress == 1 && canCompress(ext) {
		if err := compressImage(savePath, savePath, quality); err != nil {
			return c.JSON(500, map[string]string{"msg": "图片压缩失败"})
		}
	}

	// 获取最终文件大小（压缩后）
	stat, err := os.Stat(savePath)
	if err != nil {
		return c.JSON(500, map[string]string{"msg": "获取文件信息失败"})
	}
	finalSize := stat.Size()

	// 写入 DB
	_, err = db.DB.Exec(`
		INSERT INTO files(uuid, original_name, filename, size, ext, is_private)
		VALUES(?,?,?,?,?,?)`,
		uid, file.Filename, filename, finalSize, ext, isPrivate,
	)
	if err != nil {
		_ = os.Remove(savePath)
		return c.JSON(500, map[string]string{"msg": "数据库写入失败"})
	}

	return c.JSON(200, map[string]any{
		"url":      buildURL(c, filename),
		"filename": filename,
		"size":     finalSize,
	})
}
