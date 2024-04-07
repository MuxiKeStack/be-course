package service

import (
	"context"
	ccnuv1 "github.com/MuxiKeStack/be-api/gen/proto/ccnu/v1"
	"github.com/MuxiKeStack/be-course/domain"
	"github.com/MuxiKeStack/be-course/repository"
	"github.com/ecodeclub/ekit/slice"
	"golang.org/x/sync/errgroup"
	"time"
)

type CourseService interface {
	// List 这里的 uid 作为变长参数作用是这个 uid 是可选的，只有装饰容错的时候才需要传入一个 uid
	List(ctx context.Context, studentId string, password string, year string, term string, uid ...int64) ([]domain.Course, error)
	GetDetailById(ctx context.Context, id int64) (domain.Course, error)
	GetGradesById(ctx context.Context, id int64) ([]domain.Grade, error)
	FindIdOrCreateByCourse(ctx context.Context, course domain.Course) (int64, error)
}

type courseService struct {
	ccnu ccnuv1.CCNUServiceClient
	repo repository.CourseRepository
}

func (s *courseService) FindIdOrCreateByCourse(ctx context.Context, course domain.Course) (int64, error) {
	// TODO 这个方法预期频率高，要优化性能
	id, err := s.repo.FindIdByCourse(ctx, course)
	if err == nil {
		return id, nil
	}
	if err != repository.ErrCourseNotFound {
		// 系统错误，向上抛
		return 0, err
	}
	// 没找到，创建
	err = s.repo.Create(ctx, course)
	if err != nil && err != repository.ErrCourseDuplicate {
		// 系统错误
		return 0, err
	}
	// 查主库
	return s.repo.FindIdByCourse(ctx, course)
}

func (s *courseService) List(ctx context.Context, studentId string, password string, year string, term string, uid ...int64) ([]domain.Course, error) {
	// 教务系统查询，
	res, err := s.ccnu.CourseList(ctx, &ccnuv1.CourseListRequest{
		StudentId: studentId,
		Password:  password,
		Year:      year,
		Term:      term,
	})
	if err != nil {
		return nil, err
	}
	var courses []domain.Course
	courses = slice.Map(res.Courses, func(idx int, src *ccnuv1.Course) domain.Course {
		return domain.Course{
			CourseCode: src.GetCourseCode(),
			Name:       src.GetName(),
			Teacher:    src.GetTeacher(),
			School:     src.GetSchool(),
			Property:   domain.CoursePropertyFromStr(src.GetProperty()),
			Credit:     src.GetCredit(),
		}
	})
	// 要在这里聚合出courseId
	var eg errgroup.Group
	for i := range courses {
		eg.Go(func() error {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
			defer cancel()
			courses[i].Id, err = s.FindIdOrCreateByCourse(ctx, courses[i])
			return err
		})
	}
	err = eg.Wait()
	if err != nil {
		return nil, err
	}
	return courses, err
}

func (s *courseService) GetDetailById(ctx context.Context, id int64) (domain.Course, error) {
	return s.repo.FindById(ctx, id)
}

func (s *courseService) GetGradesById(ctx context.Context, id int64) ([]domain.Grade, error) {
	return s.repo.FindGradesById(ctx, id)
}
