package repository

import (
	"context"
	"github.com/MuxiKeStack/be-course/domain"
	"github.com/MuxiKeStack/be-course/pkg/logger"
	"github.com/MuxiKeStack/be-course/repository/cache"
	"github.com/MuxiKeStack/be-course/repository/dao"
	"github.com/ecodeclub/ekit/slice"
	"time"
)

type CourseSubscriptionRepository interface {
	BatchCreateCourseSubscription(ctx context.Context, cs []domain.CourseSubscription) error
	FindSubscriberUidsByCourseId(ctx context.Context, courseId int64, curUid int64, limit int64) ([]int64, error)
	FindByUidYearTermAlive(ctx context.Context, uid int64, year string, term string,
		ttl time.Duration) ([]domain.CourseSubscription, error)
}

type CachedCourseSubscriptionRepository struct {
	dao   dao.CourseSubscriptionDAO
	cache cache.CourseSubscriptionCache
	l     logger.Logger
}

func NewCachedCourseSubscriptionRepository(dao dao.CourseSubscriptionDAO, cache cache.CourseSubscriptionCache,
	l logger.Logger) CourseSubscriptionRepository {
	return &CachedCourseSubscriptionRepository{dao: dao, cache: cache, l: l}
}

func (repo *CachedCourseSubscriptionRepository) FindByUidYearTermAlive(ctx context.Context, uid int64, year string, term string,
	ttl time.Duration) ([]domain.CourseSubscription, error) {
	css, err := repo.dao.FindByUidYearTermAlive(ctx, uid, year, term, ttl)
	return slice.Map(css, func(idx int, src dao.CourseSubscription) domain.CourseSubscription {
		return repo.toSubscriptionDomain(src)
	}), err
}

func (repo *CachedCourseSubscriptionRepository) BatchCreateCourseSubscription(ctx context.Context, cs []domain.CourseSubscription) error {
	return repo.dao.BatchInsertCourseSubscription(ctx, slice.Map(cs, func(idx int, src domain.CourseSubscription) dao.CourseSubscription {
		return dao.CourseSubscription{
			Uid:      src.Uid,
			Year:     src.Year,
			Term:     src.Term,
			CourseId: src.Course.Id,
		}
	}))
}

func (repo *CachedCourseSubscriptionRepository) FindSubscriberUidsByCourseId(ctx context.Context, courseId int64, curUid int64, limit int64) ([]int64, error) {
	// TODO 这个功能，也不清楚会不会时候高频的，上线前要在这里埋点，看看频率高不高，
	// 是否需要缓存（目前感受不到有什么依据来缓存哪一部分课程的uid，非要缓存的话可以缓存第一页的，相对频率会高一些)
	if curUid == 0 && limit == 10 {
		// 查第一页，先查缓存
		res, err := repo.cache.GetFirstPageUids(ctx, courseId)
		if err == nil {
			return res, nil
		}
		if err != cache.ErrKeyNotExist {
			repo.l.Error("查询缓存课程 Subscriber 失败",
				logger.Int64("courseId", courseId), logger.Error(err))
		}
	}
	// TODO 从这里想到，如果某个课程被发起了提问，就可以预热一天 ，这个要在邀请接口里面实现
	res, err := repo.dao.FindSubscriberUidsByCourseId(ctx, courseId, curUid, limit)
	if err != nil {
		return nil, err
	}
	// 异步缓存
	go func() {
		if curUid == 0 && limit == 10 {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			// 这里进行缓存的理由是：
			// 1. 第一页本来就是频率最高的，大多数人只第一页
			// 2. 既然这个课程被一个用户提问了，说明有一些令人关注的地方，可能也被其他用户提问
			// 所有为了让更多请求命中缓存，在这里设置第一页的缓存
			// 不过上面全是臆想的场景，乐
			er := repo.cache.SetFirstPageUids(ctx, courseId, res)
			if er != nil {
				repo.l.Error("回写缓存课程 Subscriber 失败", logger.Int64("courseId", courseId), logger.Error(err))
			}
		}
	}()
	return res, nil
}

func (repo *CachedCourseSubscriptionRepository) toSubscriptionDomain(cs dao.CourseSubscription) domain.CourseSubscription {
	return domain.CourseSubscription{
		Course: domain.Course{
			Id: cs.CourseId,
		},
		Uid:  cs.Uid,
		Year: cs.Year,
		Term: cs.Term,
	}
}
