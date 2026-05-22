package main

import (
	"myoss/config"
	"myoss/internal/db"
	"myoss/internal/handler"
	mymiddleware "myoss/internal/middleware"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func init() {
	if err := config.InitConfig(); err != nil {
		panic("配置加载失败：" + err.Error())
	}

	if err := db.InitDB(config.Config.DB.DSN); err != nil {
		panic(err)
	}

	err := os.MkdirAll(handler.GetUploadDir(), 0755)
	if err != nil {
		panic(err)
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.BodyLimit(50 << 20))

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAuthorization},
	}))

	e.GET("/i/:filename", handler.ServeFile)

	api := e.Group("/api", mymiddleware.ApiAuth)
	api.POST("/upload", handler.Upload)
	api.GET("/files", handler.ListFiles)
	api.DELETE("/file/:id", handler.DeleteFile)

	e.Start(config.Config.Server.Port)
}
