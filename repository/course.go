package repository

import (
	"context"
	coursev1 "github.com/MuxiKeStack/be-api/gen/proto/course/v1"
	"github.com/MuxiKeStack/be-course/domain"
	"github.com/MuxiKeStack/be-course/repository/cache"
	"github.com/MuxiKeStack/be-course/repository/dao"
)

var (
	ErrCourseNotFound  = dao.ErrRecordNorFound
	ErrCourseDuplicate = dao.ErrCourseDuplicate
)

type CourseRepository interface {
	FindById(ctx context.Context, id int64) (domain.Course, error)
	FindIdByCourse(ctx context.Context, course domain.Course) (int64, error)
	Create(ctx context.Context, course domain.Course) error
	Upsert(ctx context.Context, course domain.Course) error
	FindIdByCourseWithoutUnknownProperty(ctx context.Context, course domain.Course) (int64, error)
}

type CachedCourseRepository struct {
	dao   dao.CourseDAO
	cache cache.CourseCache
}

func NewCachedCourseRepository(dao dao.CourseDAO, cache cache.CourseCache) CourseRepository {
	return &CachedCourseRepository{dao: dao, cache: cache}
}

func (repo *CachedCourseRepository) Upsert(ctx context.Context, course domain.Course) error {
	return repo.dao.Upsert(ctx, repo.ToEntity(course))
}

func (repo *CachedCourseRepository) FindIdByCourse(ctx context.Context, course domain.Course) (int64, error) {
	// 这里考虑到一个课程，就算缓存了id，只有上过这门课的学生会调这个接口
	// 对于一门课，在校生里面最多几百人，比如一门通核8、90人，教了三届，也就不到300人
	// 而且大多数还是人更少的专业课，假设用户量能有在校生三分之一，缓存半小时的话，半小时能有几个人命中这个缓存
	// 缓存了收益很低，更何况，查询条件完全命中索引，所以不缓存了
	return repo.dao.FindIdByCourse(ctx, repo.ToEntity(course))
}

func (repo *CachedCourseRepository) FindIdByCourseWithoutUnknownProperty(ctx context.Context, course domain.Course) (int64, error) {
	return repo.dao.FindIdByCourseWithoutUnknownProperty(ctx, repo.ToEntity(course))
}

func (repo *CachedCourseRepository) Create(ctx context.Context, course domain.Course) error {
	return repo.dao.Insert(ctx, repo.ToEntity(course))
}

func (repo *CachedCourseRepository) FindById(ctx context.Context, id int64) (domain.Course, error) {
	// TODO 先查缓存，新发的课评会预热相关课程
	res, err := repo.cache.Get(ctx, id)
	if err == nil {
		return res, nil
	}
	// 1. 没有key
	// 2. redis崩溃，这里预期没有缓存也撑得住，不采取降级来保护数据库
	c, err := repo.dao.FindById(ctx, id)
	if err != nil {
		return domain.Course{}, err
	}
	return repo.ToDomain(c), err
}

func (repo *CachedCourseRepository) ToEntity(course domain.Course) dao.Course {
	return dao.Course{
		Id:         course.Id,
		CourseCode: course.CourseCode,
		Name:       course.Name,
		Teacher:    course.Teacher,
		School:     course.School,
		Property:   int32(course.Property),
		Credit:     course.Credit,
	}
}

func (repo *CachedCourseRepository) ToDomain(c dao.Course) domain.Course {
	return domain.Course{
		Id:         c.Id,
		CourseCode: c.CourseCode,
		Name:       c.Name,
		Teacher:    c.Teacher,
		School:     c.School,
		Property:   coursev1.CourseProperty(c.Property),
		Credit:     c.Credit,
	}
}
