package repository

import (
	"context"
	"github.com/MuxiKeStack/be-course/domain"
	"github.com/MuxiKeStack/be-course/repository/dao"
	"github.com/ecodeclub/ekit/slice"
)

type CourseSubscriptionRepository interface {
	BatchCreateCourseSubscription(ctx context.Context, cs []domain.CourseSubscription) error
	FindCourseIdsByUidYearTerm(ctx context.Context, uid int64, year string, term string) ([]int64, error)
}

type CachedSubscriptionRepository struct {
	dao dao.CourseSubscriptionDAO
}

func (repo *CachedSubscriptionRepository) FindCourseIdsByUidYearTerm(ctx context.Context, uid int64, year string, term string) ([]int64, error) {
	return repo.dao.FindCourseIdsByUidYearTerm(ctx, uid, year, term)
}

func (repo *CachedSubscriptionRepository) BatchCreateCourseSubscription(ctx context.Context, cs []domain.CourseSubscription) error {
	return repo.dao.BatchInsertCourseSubscription(ctx, slice.Map(cs, func(idx int, src domain.CourseSubscription) dao.CourseSubscription {
		return dao.CourseSubscription{
			Uid:      src.Uid,
			Year:     src.Year,
			Term:     src.Term,
			CourseId: src.CourseId,
		}
	}))
}
