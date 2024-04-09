package service

import (
	"context"
	"github.com/MuxiKeStack/be-course/domain"
	"github.com/MuxiKeStack/be-course/pkg/logger"
	"github.com/MuxiKeStack/be-course/repository"
	"time"
)

type PerformanceCourseService struct {
	CourseService
	repo        repository.CourseRepository
	currentYear string
	currentTerm string
	selecting   bool // 选课中，配置文件中手动配置
	courseTTL   time.Duration
	l           logger.Logger
}

func NewPerformanceCourseService(courseService CourseService, repo repository.CourseRepository,
	currentYear string, currentTerm string, selecting bool, courseTTL time.Duration, l logger.Logger) *PerformanceCourseService {
	return &PerformanceCourseService{CourseService: courseService, repo: repo, currentYear: currentYear, currentTerm: currentTerm, selecting: selecting, courseTTL: courseTTL, l: l}
}

func (p *PerformanceCourseService) List(ctx context.Context, studentId string, password string, year string, term string, uid ...int64) ([]domain.Course, error) {
	// 这里为了提高性能，加一个装饰器， 如果在课程稳定时间段：历史学年期，非选课时间内，直接拦一下，看数据库有没较新的数据，有的话直接返回，不再去爬取了
	if len(uid) == 0 {
		return nil, ErrUidNotInput
	}
	isHistory := year < p.currentYear || year == p.currentYear && term < p.currentTerm
	isStable := isHistory || !p.selecting // 是否在课程稳定时间段
	if isStable {
		// 去数据库看
		courses, err := p.repo.FindByUidYearTermAlive(ctx, uid[0], year, term, p.courseTTL)
		if err != nil {
			p.l.Error("从数据库获取课程失败", logger.Error(err))
		}
		if len(courses) != 0 {
			return courses, nil
		}
	}
	return p.CourseService.List(ctx, studentId, password, year, term, uid...)
}
