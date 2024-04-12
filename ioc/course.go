package ioc

import (
	ccnuv1 "github.com/MuxiKeStack/be-api/gen/proto/ccnu/v1"
	"github.com/MuxiKeStack/be-course/event"
	"github.com/MuxiKeStack/be-course/pkg/logger"
	"github.com/MuxiKeStack/be-course/repository"
	"github.com/MuxiKeStack/be-course/service"
	"github.com/spf13/viper"
	"time"
)

func InitFallBackCourseService(ccnu ccnuv1.CCNUServiceClient, repo repository.CourseRepository,
	producer event.Producer, l logger.Logger, subRepo repository.CourseSubscriptionRepository) service.CourseService {
	type Config struct {
		Year   string `yaml:"year"`
		Term   string `yaml:"term"`
		Course struct {
			Selecting bool  `yaml:"selecting"`
			TTL       int64 `yaml:"TTL"`
		} `yaml:"course"`
	}
	var cfg *Config
	err := viper.UnmarshalKey("current", &cfg)
	if err != nil {
		panic(err)
	}
	courseService := service.NewCourseService(ccnu, repo, subRepo, cfg.Year, cfg.Term)
	fc := service.NewFallbackCourseService(courseService, repo, producer, l)
	return fc
}

func InitPerformanceFallBackCourseService(ccnu ccnuv1.CCNUServiceClient, repo repository.CourseRepository,
	producer event.Producer, l logger.Logger, subRepo repository.CourseSubscriptionRepository) service.CourseService {
	type Config struct {
		Year   string `yaml:"year"`
		Term   string `yaml:"term"`
		Course struct {
			Selecting bool  `yaml:"selecting"`
			TTL       int64 `yaml:"TTL"`
		} `yaml:"course"`
	}
	var cfg *Config
	err := viper.UnmarshalKey("current", &cfg)
	if err != nil {
		panic(err)
	}
	courseService := service.NewCourseService(ccnu, repo, subRepo, cfg.Year, cfg.Term)
	fc := service.NewFallbackCourseService(courseService, repo, producer, l)
	courseTTL := time.Duration(cfg.Course.TTL) * time.Hour * 24
	//courseTTL := time.Second
	pfc := service.NewPerformanceCourseService(fc, repo, cfg.Year,
		cfg.Term, cfg.Course.Selecting, courseTTL, l)
	return pfc
}
