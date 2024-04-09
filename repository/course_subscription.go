package repository

import (
	"context"
	"github.com/MuxiKeStack/be-course/domain"
	"github.com/MuxiKeStack/be-course/repository/dao"
	"github.com/ecodeclub/ekit/slice"
	"time"
)

type CourseSubscriptionRepository interface {
	BatchCreateCourseSubscription(ctx context.Context, cs []domain.CourseSubscription) error
	FindCourseIdsByUidYearTerm(ctx context.Context, uid int64, year string, term string) ([]int64, error)
	FindCourseIdsByUidYearTermAlive(ctx context.Context, uid int64, year string, term string, TTL time.Duration) ([]int64, error)
}

type courseSubscriptionRepository struct {
	dao dao.CourseSubscriptionDAO
}

func NewCourseSubscriptionRepository(dao dao.CourseSubscriptionDAO) CourseSubscriptionRepository {
	return &courseSubscriptionRepository{dao: dao}
}

func (repo *courseSubscriptionRepository) FindCourseIdsByUidYearTerm(ctx context.Context, uid int64, year string, term string) ([]int64, error) {
	return repo.dao.FindCourseIdsByUidYearTerm(ctx, uid, year, term)
}

func (repo *courseSubscriptionRepository) FindCourseIdsByUidYearTermAlive(ctx context.Context, uid int64, year string, term string,
	TTL time.Duration) ([]int64, error) {
	return repo.dao.FindCourseIdsByUidYearTermAlive(ctx, uid, year, term, TTL)
}

func (repo *courseSubscriptionRepository) BatchCreateCourseSubscription(ctx context.Context, cs []domain.CourseSubscription) error {
	return repo.dao.BatchInsertCourseSubscription(ctx, slice.Map(cs, func(idx int, src domain.CourseSubscription) dao.CourseSubscription {
		return dao.CourseSubscription{
			Uid:      src.Uid,
			Year:     src.Year,
			Term:     src.Term,
			CourseId: src.CourseId,
		}
	}))
}
