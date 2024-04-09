package main

import (
	"github.com/MuxiKeStack/be-course/pkg/grpcx"
	"github.com/MuxiKeStack/be-course/pkg/saramax"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	initViper()
	app := InitApp()
	for _, c := range app.consumers {
		err := c.Start()
		if err != nil {
			panic(err)
		}
	}
	err := app.server.Serve()
	if err != nil {
		panic(err)
	}

}

func initViper() {
	cfile := pflag.String("config", "config/config.yaml", "配置文件路径")
	pflag.Parse()

	viper.SetConfigType("yaml")
	viper.SetConfigFile(*cfile)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

type App struct {
	server    grpcx.Server
	consumers []saramax.Consumer
}
