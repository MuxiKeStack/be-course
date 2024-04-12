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
	FindSubscriberUidsByCourseId(ctx context.Context, courseId int64, curUid int64, limit int64) ([]int64, error)
	FindByUidYearTermAlive(ctx context.Context, uid int64, year string, term string, ttl time.Duration) ([]CourseSubscription, error)
}

type GORMCourseSubscriptionDAO struct {
	db *gorm.DB
}

func NewGORMCourseSubscriptionDAO(db *gorm.DB) CourseSubscriptionDAO {
	return &GORMCourseSubscriptionDAO{db: db}
}

func (dao *GORMCourseSubscriptionDAO) BatchInsertCourseSubscription(ctx context.Context, subscriptions []CourseSubscription) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var eg errgroup.Group
		for _, s := range subscriptions {
			eg.Go(func() error {
				s.Utime = now
				s.Ctime = now
				return tx.Clauses(
					clause.OnConflict{DoUpdates: clause.Assignments(map[string]any{
						"utime": now,
					})}).Create(&s).Error
			})
		}
		return eg.Wait()
	})
}

func (dao *GORMCourseSubscriptionDAO) FindByUidYearTermAlive(ctx context.Context, uid int64, year string,
	term string, TTL time.Duration) ([]CourseSubscription, error) {
	query := dao.db.WithContext(ctx).
		Where("uid = ? and utime > ?", uid, time.Now().Add(-TTL).UnixMilli())
	if year != "" {
		query.Where("year = ?", year)
	}
	if term != "" {
		query.Where("term = ?", term)
	}
	var cs []CourseSubscription
	err := query.Find(&cs).Error
	return cs, err
}

func (dao *GORMCourseSubscriptionDAO) FindSubscriberUidsByCourseId(ctx context.Context, courseId int64,
	curUid int64, limit int64) ([]int64, error) {
	var uids []int64
	// 这里以uid作为 offset 的话就要让uid是有序的，这样有一个不好的地方，一个专业的人大部分课一样，以这些课为单位的顺序
	// 如果以天然有序的id作为存续，那么就要做返回一个结果，这样其实也和上面差不多，懒得这样干了
	// 不需要是没有评过的因为是让他来回答我的问题，目前这样就可以
	err := dao.db.WithContext(ctx).
		Model(&CourseSubscription{}).
		Select("uid").
		Where("course_id = ? and uid > ?", courseId, curUid).
		Order("uid asc").
		Limit(int(limit)).Find(&uids).Error
	return uids, err
}

type CourseSubscription struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 下面四个是高频查询字段，需要设置索引加速查询，作为前缀
	Uid  int64  `gorm:"uniqueIndex:uid_year_term_courseId"`
	Year string `gorm:"uniqueIndex:uid_year_term_courseId; type:char(4)"`
	Term string `gorm:"uniqueIndex:uid_year_term_courseId; type:char(1)"`
	// course_id 和其他字段组合的结果需要时唯一的，所以要放在尾部
	CourseId int64 `gorm:"uniqueIndex:uid_year_term_courseId"`
	Utime    int64 // 这里历史查询条件，但是特地为utime建立索引感觉没太大必要，因为前面的条件已经把大多数行筛掉了
	Ctime    int64
}
