package repository

import (
	"context"
	"github.com/MuxiKeStack/be-course/domain"
	"github.com/MuxiKeStack/be-course/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

var (
	ErrCourseNotFound  = dao.ErrRecordNorFound
	ErrCourseDuplicate = dao.ErrCourseDuplicate
)

type CourseRepository interface {
	FindById(ctx context.Context, id int64) (domain.Course, error)
	FindGradesById(ctx context.Context, id int64) ([]domain.Grade, error)
	FindByUidYearTerm(ctx context.Context, uid int64, year string, term string) ([]domain.Course, error)
	FindIdByCourse(ctx context.Context, course domain.Course) (int64, error)
	Create(ctx context.Context, course domain.Course) error
}

type CachedCourseRepository struct {
	dao     dao.CourseDAO
	subRepo CourseSubscriptionRepository
}

func (repo *CachedCourseRepository) FindIdByCourse(ctx context.Context, course domain.Course) (int64, error) {
	return repo.dao.FindIdByCourse(ctx, repo.courseToEntity(course))
}

func (repo *CachedCourseRepository) Create(ctx context.Context, course domain.Course) error {
	return repo.dao.Insert(ctx, repo.courseToEntity(course))
}

func (repo *CachedCourseRepository) FindByUidYearTerm(ctx context.Context, uid int64, year string, term string) ([]domain.Course, error) {
	// 不使用join语句，分步：先拿到courseIds
	cids, err := repo.subRepo.FindCourseIdsByUidYearTerm(ctx, uid, year, term)
	if err != nil {
		return nil, err
	}
	cs, err := repo.dao.FindByIds(ctx, cids)
	return slice.Map(cs, func(idx int, src dao.Course) domain.Course {
		return repo.courseToDomain(src)
	}), err
}

func (repo *CachedCourseRepository) FindById(ctx context.Context, id int64) (domain.Course, error) {
	c, err := repo.dao.FindById(ctx, id)
	return repo.courseToDomain(c), err
}

func (repo *CachedCourseRepository) FindGradesById(ctx context.Context, id int64) ([]domain.Grade, error) {
	gs, err := repo.dao.FindGradesById(ctx, id)
	return slice.Map(gs, func(idx int, src dao.Grade) domain.Grade {
		return repo.gradeToDomain(src)
	}), err
}

func (repo *CachedCourseRepository) courseToEntity(course domain.Course) dao.Course {
	return dao.Course{
		Id:         course.Id,
		CourseCode: course.CourseCode,
		Name:       course.Name,
		Teacher:    course.Teacher,
		School:     course.School,
		Property:   course.Property.Uint8(),
		Credit:     course.Credit,
	}
}

func (repo *CachedCourseRepository) courseToDomain(c dao.Course) domain.Course {
	return domain.Course{
		Id:         c.Id,
		CourseCode: c.CourseCode,
		Name:       c.Name,
		Teacher:    c.Teacher,
		School:     c.School,
		Property:   domain.CourseProperty(c.Property),
		Credit:     c.Credit,
	}
}

func (repo *CachedCourseRepository) gradeToDomain(g dao.Grade) domain.Grade {
	return domain.Grade{
		Regular: g.Regular,
		Final:   g.Final,
		Total:   g.Total,
		Year:    g.Year,
		Term:    g.Term,
	}
}
