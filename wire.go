//go:build wireinject

package main

import (
	"github.com/MuxiKeStack/be-course/event"
	"github.com/MuxiKeStack/be-course/grpc"
	"github.com/MuxiKeStack/be-course/ioc"
	"github.com/MuxiKeStack/be-course/repository"
	"github.com/MuxiKeStack/be-course/repository/cache"
	"github.com/MuxiKeStack/be-course/repository/dao"
	"github.com/google/wire"
)

func InitApp() *App {
	wire.Build(
		wire.Struct(new(App), "*"),
		//consumer
		ioc.InitConsumers,
		event.NewCourseListEventConsumer,
		// grpc
		ioc.InitGRPCxKratosServer,
		grpc.NewCourseServiceServer,
		ioc.InitPerformanceFallBackCourseService,
		ioc.InitProducer,
		ioc.InitKafka,
		repository.NewCachedCourseRepository, repository.NewCachedCourseSubscriptionRepository,
		cache.NewRedisCourseCache, cache.NewRedisCourseSubscriptionCache,
		dao.NewGORMCourseDAO, dao.NewGORMCourseSubscriptionDAO,
		ioc.InitCCNUClient,
		// 第三方组件
		ioc.InitRedis,
		ioc.InitEtcdClient,
		ioc.InitDB,
		ioc.InitLogger,
	)
	return &App{}
}
