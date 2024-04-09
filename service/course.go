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
	ccnu        ccnuv1.CCNUServiceClient
	repo        repository.CourseRepository
	currentYear string
	currentTerm string
}

func NewCourseService(ccnu ccnuv1.CCNUServiceClient, repo repository.CourseRepository,
	currentYear string, currentTerm string) CourseService {
	return &courseService{ccnu: ccnu, repo: repo,
		currentYear: currentYear, currentTerm: currentTerm}
}

func (s *courseService) FindIdOrCreateByCourse(ctx context.Context, course domain.Course) (int64, error) {
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
	// 从课程接口，判断是否选课中，好像都无所谓，都返回就行了，但是后面发表课评的时候要判断是否选课中
	isHistory := year < s.currentYear || year == s.currentYear && term < s.currentTerm
	var src ccnuv1.Source
	// 判断学年期，从成绩接口还是老接口
	if isHistory {
		src = ccnuv1.Source_GradeApi
	} else {
		src = ccnuv1.Source_OldXkApi
	}
	var courses []domain.Course
	res, err := s.ccnu.CourseList(ctx, &ccnuv1.CourseListRequest{
		StudentId: studentId,
		Password:  password,
		Year:      year,
		Term:      term,
		Source:    src,
	})
	if err != nil {
		return nil, err
	}
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

	// 要在这里聚合出courseId，两种查询结果要采用不同的聚合手段,两个不同的聚合id的接口	[优胜劣汰]
	var eg errgroup.Group
	if isHistory {
		for i := range courses {
			eg.Go(func() error {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
				defer cancel()
				courses[i].Id, err = s.FindIdOrUpsertByCourse(ctx, courses[i])
				return err
			})
		}
	} else {
		for i := range courses {
			eg.Go(func() error {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
				defer cancel()
				courses[i].Id, err = s.FindIdOrCreateByCourse(ctx, courses[i])
				return err
			})
		}
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

func (s *courseService) FindIdOrUpsertByCourse(ctx context.Context, course domain.Course) (int64, error) {
	id, err := s.repo.FindIdByCourseWithoutUnknownProperty(ctx, course)
	if err == nil {
		return id, nil
	}
	if err != repository.ErrCourseNotFound {
		// 系统错误，向上抛
		return 0, err
	}
	// 没找到，创建
	err = s.repo.Upsert(ctx, course)
	if err != nil {
		// 系统错误
		return 0, err
	}
	// 查主库
	return s.repo.FindIdByCourse(ctx, course)
}
