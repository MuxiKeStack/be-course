package service

import (
	"context"
	"errors"
	ccnuv1 "github.com/MuxiKeStack/be-api/gen/proto/ccnu/v1"
	"github.com/MuxiKeStack/be-course/domain"
	"github.com/MuxiKeStack/be-course/event"
	"github.com/MuxiKeStack/be-course/pkg/logger"
	"github.com/MuxiKeStack/be-course/repository"
)

var (
	ErrDownGradeCourseNotFound = errors.New("降级课程未找到")
	ErrUidNotInput             = errors.New("用户id未传入")
)

// FallbackCourseService 故障转移装饰层
type FallbackCourseService struct {
	CourseService
	repo     repository.CourseRepository
	producer event.Producer
	l        logger.Logger
}

func (f *FallbackCourseService) List(ctx context.Context, studentId string, password string, year string, term string, uid ...int64) ([]domain.Course, error) {
	if len(uid) == 0 {
		return nil, ErrUidNotInput
	}
	courses, err := f.CourseService.List(ctx, studentId, password, year, term)
	switch {
	case err == nil:
		// 开kafka异步存入数据库
		for _, c := range courses {
			er := f.producer.ProduceCourseListEvent(context.Background(), event.CourseFromXkEvent{
				CourseId: c.Id,
				Uid:      uid[0],
				Year:     year,
				Term:     term,
			})
			if er != nil {
				f.l.Error("生产CourseListEvent失败", logger.Error(err), logger.String("studentId", studentId))
			}
		}
	case ccnuv1.IsNetworkToXkError(err):
		// 降级,从数据查旧的数据，没查到就直接返回
		var er error
		courses, er = f.repo.FindByUidYearTerm(ctx, uid[0], year, term)
		if er != nil {
			return nil, er
		}
		if len(courses) == 0 {
			return nil, ErrDownGradeCourseNotFound
		}
	}
	return courses, err
}
