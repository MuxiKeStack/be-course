package dao

import (
	"context"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type CourseSubscriptionDAO interface {
	BatchInsertCourseSubscription(ctx context.Context, subscriptions []CourseSubscription) error
	FindCourseIdsByUidYearTerm(ctx context.Context, uid int64, year string, term string) ([]int64, error)
}

type GORMCourseSubscriptionDAO struct {
	db *gorm.DB
}

func (dao *GORMCourseSubscriptionDAO) BatchInsertCourseSubscription(ctx context.Context, subscriptions []CourseSubscription) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var eg errgroup.Group
		for _, s := range subscriptions {
			eg.Go(func() error {
				return tx.Clauses(
					clause.OnConflict{DoUpdates: clause.Assignments(map[string]any{
						"utime": now,
					})}).Create(&s).Error
			})
		}
		return eg.Wait()
	})
}

func (dao *GORMCourseSubscriptionDAO) FindCourseIdsByUidYearTerm(ctx context.Context, uid int64, year string, term string) ([]int64, error) {
	var ids []int64
	err := dao.db.WithContext(ctx).
		Model(&CourseSubscription{}).
		Select("course_id").
		Where("uid = ? and year = ? and term = ?", uid, year, term).
		Find(&ids).Error
	return ids, err
}

type CourseSubscription struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 下面三个是高频查询字段，需要设置索引加速查询，作为前缀
	Uid  int64  `gorm:"uniqueIndex:uid_year_term_courseId"`
	Year string `gorm:"uniqueIndex:uid_year_term_courseId"`
	Term string `gorm:"uniqueIndex:uid_year_term_courseId"`
	// course_id 和其他字段组合的结果需要时唯一的，所以要放在尾部
	CourseId int64 `gorm:"uniqueIndex:uid_year_term_courseId"`
	Utime    int64
	Ctime    int64
}