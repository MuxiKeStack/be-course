package repository

import (
	"context"
	"github.com/MuxiKeStack/be-course/domain"
	"github.com/MuxiKeStack/be-course/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type CourseRepository interface {
	BatchCreate(ctx context.Context, courses []domain.FailOverCourse) error
	FindByStudentIdYearTerm(ctx context.Context, studentId string, year string, term string) ([]domain.FailOverCourse, error)
}

type CachedCourseRepository struct {
	dao dao.CourseDAO
}

func (repo *CachedCourseRepository) BatchCreate(ctx context.Context, courses []domain.FailOverCourse) error {
	return repo.dao.BatchInsert(ctx, slice.Map(courses, func(idx int, src domain.FailOverCourse) dao.FailOverCourse {
		return repo.toEntity(src)
	}))
}

func (repo *CachedCourseRepository) FindByStudentIdYearTerm(ctx context.Context, studentId string,
	year string, term string) ([]domain.FailOverCourse, error) {
	fcs, err := repo.dao.FindByStudentIdYearTerm(ctx, studentId, year, term)
	return slice.Map(fcs, func(idx int, src dao.FailOverCourse) domain.FailOverCourse {
		return repo.toDomain(src)
	}), err
}

func (repo *CachedCourseRepository) toEntity(course domain.FailOverCourse) dao.FailOverCourse {
	return dao.FailOverCourse{
		Id:        course.Id,
		StudentId: course.StudentId,
		Year:      course.Year,
		Term:      course.Term,
		CourseId:  course.CourseId,
		Name:      course.Name,
		Teacher:   course.Teacher,
	}
}

func (repo *CachedCourseRepository) toDomain(course dao.FailOverCourse) domain.FailOverCourse {
	return domain.FailOverCourse{
		Id:        course.Id,
		StudentId: course.StudentId,
		CourseId:  course.CourseId,
		Name:      course.Name,
		Teacher:   course.Teacher,
		Year:      course.Year,
		Term:      course.Term,
	}
}
