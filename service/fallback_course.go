package service

import (
	"context"
	ccnuv1 "github.com/MuxiKeStack/be-api/gen/proto/ccnu/v1"
	"github.com/MuxiKeStack/be-course/domain"
	"github.com/MuxiKeStack/be-course/repository"
)

type FallbackCourseService struct {
	CourseService
	repo repository.CourseRepository
}

func (f *FallbackCourseService) List(ctx context.Context, studentId string, password string, year string, term string) ([]domain.FailoverCourse, error) {
	courses, err := f.CourseService.List(ctx, studentId, password, year, term)
	if ccnuv1.IsNetworkToXkError(err) {
		// 降级,从数据查旧的数据，没查到就直接返回
		var er error
		courses, er = f.repo.FindByStudentIdYearTerm(ctx, studentId, year, term)
		if er != nil {
			return nil, er
		}
		if len(courses) == 0 {
			return nil, ErrDownGradeCourseNotFound
		}
	}
	return courses, err
}
