package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/quarkcms/quark-go/examples/zeroadmin/internal/config"
	"github.com/quarkcms/quark-go/examples/zeroadmin/internal/handler"
	"github.com/quarkcms/quark-go/examples/zeroadmin/internal/svc"
	"github.com/quarkcms/quark-go/pkg/adapter/zeroadapter"
	"github.com/quarkcms/quark-go/pkg/app/handler/admin"
	"github.com/quarkcms/quark-go/pkg/app/install"
	"github.com/quarkcms/quark-go/pkg/app/middleware"
	"github.com/quarkcms/quark-go/pkg/builder"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/zeroadmin-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// 加载静态文件
	staticFile("/", "./website", server)

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	// 数据库配置信息
	dsn := "root:Bc5HQFJc4bLjZCcC@tcp(127.0.0.1:3306)/quarkgo?charset=utf8&parseTime=True&loc=Local"

	// 配置资源
	config := &builder.Config{
		AppKey:    "123456",
		Providers: admin.Providers,
		DBConfig: &builder.DBConfig{
			Dialector: mysql.Open(dsn),
			Opts:      &gorm.Config{},
		},
	}

	// 创建对象
	b := builder.New(config)

	// 初始化安装
	b.Use(install.Handle)

	// 中间件
	b.Use(middleware.Handle)

	// 适配gozero
	zeroadapter.Adapter(b, server)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

// 加载静态文件
func staticFile(root string, dirPath string, server *rest.Server) {
	rd, _ := ioutil.ReadDir(dirPath)

	for _, f := range rd {
		fileName := f.Name()
		subPath := root + fileName + "/"
		subDirPath := dirPath + "/" + fileName
		if isDir(subDirPath) {
			staticFile(subPath, subDirPath, server)
		}
	}

	server.AddRoute(
		rest.Route{
			Method:  http.MethodGet,
			Path:    root + ":file",
			Handler: http.StripPrefix(root, http.FileServer(http.Dir(dirPath))).ServeHTTP,
		},
	)
}

// 判断所给路径是否为文件夹
func isDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}
