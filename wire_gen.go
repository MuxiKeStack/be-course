// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/MuxiKeStack/be-course/event"
	"github.com/MuxiKeStack/be-course/grpc"
	"github.com/MuxiKeStack/be-course/ioc"
	"github.com/MuxiKeStack/be-course/repository"
	"github.com/MuxiKeStack/be-course/repository/cache"
	"github.com/MuxiKeStack/be-course/repository/dao"
)

// Injectors from wire.go:

func InitApp() *App {
	client := ioc.InitEtcdClient()
	ccnuServiceClient := ioc.InitCCNUClient(client)
	logger := ioc.InitLogger()
	db := ioc.InitDB(logger)
	courseDAO := dao.NewGORMCourseDAO(db)
	cmdable := ioc.InitRedis()
	courseCache := cache.NewRedisCourseCache(cmdable)
	courseRepository := repository.NewCachedCourseRepository(courseDAO, courseCache)
	saramaClient := ioc.InitKafka()
	producer := ioc.InitProducer(saramaClient)
	courseSubscriptionDAO := dao.NewGORMCourseSubscriptionDAO(db)
	courseSubscriptionCache := cache.NewRedisCourseSubscriptionCache(cmdable)
	courseSubscriptionRepository := repository.NewCachedCourseSubscriptionRepository(courseSubscriptionDAO, courseSubscriptionCache, logger)
	courseService := ioc.InitPerformanceFallBackCourseService(ccnuServiceClient, courseRepository, producer, logger, courseSubscriptionRepository)
	courseServiceServer := grpc.NewCourseServiceServer(courseService)
	server := ioc.InitGRPCxKratosServer(courseServiceServer, client, logger)
	courseListEventConsumer := event.NewCourseListEventConsumer(saramaClient, logger, courseSubscriptionRepository)
	v := ioc.InitConsumers(courseListEventConsumer)
	app := &App{
		server:    server,
		consumers: v,
	}
	return app
}
