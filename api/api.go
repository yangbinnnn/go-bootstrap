package api

import (
	"io"
	"os"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"

	"go-box/common"
	mymid "go-box/middleware"
)

var (
	e *echo.Echo
)

func StartAPP() {
	config := common.GlobalConfig
	e.Logger.Fatal(e.Start(config.HTTPAddr))
}

func InitApi() {
	config := common.GlobalConfig
	logfile, _ := os.OpenFile(config.LogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	tee := io.MultiWriter(os.Stdout, logfile)
	logconfig := middleware.LoggerConfig{
		Output: tee,
	}

	e = echo.New()
	e.Logger.SetLevel(log.DEBUG)

	// middlewares
	mymid.InitMiddleware()
	e.Use(mymid.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.CORS())
	e.Use(middleware.LoggerWithConfig(logconfig))

	// web
	e.File("/", config.WebIndex)
	e.Static("/static", config.WebStatic)

	// docs
	if config.Debug {
		e.Static("/docs", config.DocStatic)
	}

	// ws
	e.GET("/ws", common.WSRegistry)

	// api
	srv := e.Group(config.APIPrefix)
	if config.Debug {
		// record the json request/response for debug
		jsonBodyConfig := middleware.BodyDumpConfig{
			Skipper: mymid.JsonBodySkipper,
			Handler: mymid.JsonBodyHandler,
		}
		srv.Use(middleware.BodyDumpWithConfig(jsonBodyConfig))
	}
	srv.GET("/ping", ping)

	// api math
	math := srv.Group("/math")
	math.GET("/add", add)

	// api user
	user := srv.Group("/user")
	user.POST("/register", userRegister)
	user.POST("/login", userLogin)

	// token auth
	user.GET("/info", userInfo, mymid.TokenAuth.Process)

	// api file
	file := srv.Group("/file")
	file.GET("/:name", download)
	file.PUT("/:name", upload)
}
