package service

import (
	"context"
	"errors"
	ccnuv1 "github.com/MuxiKeStack/be-api/gen/proto/ccnu/v1"
	"github.com/MuxiKeStack/be-course/domain"
	"github.com/MuxiKeStack/be-course/event"
	"github.com/MuxiKeStack/be-course/pkg/logger"
	"github.com/MuxiKeStack/be-course/repository"
	"github.com/ecodeclub/ekit/slice"
)

var ErrDownGradeCourseNotFound = errors.New("降级课程未找到")

type CourseService interface {
	List(ctx context.Context, studentId string, password string, year string, term string) ([]domain.FailOverCourse, error)
}

type courseService struct {
	ccnu     ccnuv1.CCNUServiceClient
	producer event.Producer
	repo     repository.CourseRepository
	l        logger.Logger
}

func (s *courseService) List(ctx context.Context, studentId string, password string, year string, term string) ([]domain.FailOverCourse, error) {
	// 教务系统查询，
	// TODO 这里ccnu服务要failover装饰
	res, err := s.ccnu.CourseList(ctx, &ccnuv1.CourseListRequest{
		StudentId: studentId,
		Password:  password,
		Year:      year,
		Term:      term,
	})
	if err != nil && !ccnuv1.IsNetworkToXkError(err) {
		return nil, err
	}

	var courses []domain.FailOverCourse
	if err == nil {
		courses = slice.Map(res.Courses, func(idx int, src *ccnuv1.Course) domain.FailOverCourse {
			return domain.FailOverCourse{
				CourseId: src.CourseId,
				Name:     src.Name,
				Teacher:  src.Teacher,
				Year:     src.Year,
				Term:     src.Term,
			}
		})
		// 开kafka异步存入数据库
		er := s.producer.ProduceCourseListEvent(context.Background(), event.CourseListEvent{
			Courses: courses,
		})
		if er != nil {
			s.l.Error("生产CourseListEvent失败", logger.Error(err), logger.String("studentId", studentId))
		}
	} else {
		// 降级,从数据查旧的数据，没查到就直接返回
		var er error
		courses, er = s.repo.FindByStudentIdYearTerm(ctx, studentId, year, term)
		if er != nil {
			return nil, er
		}
		if len(courses) == 0 {
			return nil, ErrDownGradeCourseNotFound
		}
	}
	return courses, err
}
