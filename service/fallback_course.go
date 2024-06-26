package service

import (
	"context"
	"errors"
	ccnuv1 "github.com/MuxiKeStack/be-api/gen/proto/ccnu/v1"
	"github.com/MuxiKeStack/be-course/domain"
	"github.com/MuxiKeStack/be-course/event"
	"github.com/MuxiKeStack/be-course/pkg/logger"
	"github.com/MuxiKeStack/be-course/repository"
)

var (
	ErrDownGradeCourseNotFound = errors.New("降级课程未找到")
	ErrUidNotInput             = errors.New("用户id未传入")
)

// FallbackCourseService 降级装饰层
type FallbackCourseService struct {
	CourseService
	repo        repository.CourseRepository
	producer    event.Producer
	l           logger.Logger
	currentYear string
	currentTerm string
	selecting   bool // 选课中，配置文件中手动配置
}

func NewFallbackCourseService(courseService CourseService, repo repository.CourseRepository,
	producer event.Producer, l logger.Logger, currentYear string, currentTerm string, selecting bool) CourseService {
	return &FallbackCourseService{
		CourseService: courseService,
		repo:          repo,
		producer:      producer,
		l:             l,
		currentYear:   currentYear,
		currentTerm:   currentTerm,
		selecting:     selecting,
	}
}

func (f *FallbackCourseService) SubscriptionList(ctx context.Context, studentId string, password string, year string,
	term string, uid ...int64) ([]domain.CourseSubscription, error) {
	// 这里为了降级，加一个装饰器
	if len(uid) == 0 {
		return nil, ErrUidNotInput
	}
	courseSubscriptions, err := f.CourseService.SubscriptionList(ctx, studentId, password, year, term)
	switch {
	case err == nil:
		// 查询的课程不是在选课时间段的课程，开kafka异步存入数据库，这样可以保证，在数据库subscribed的课程都是可以评价的选上的课程
		isHistory := year < f.currentYear || year == f.currentYear && term < f.currentTerm
		isStable := isHistory || !f.selecting // 是否在课程稳定时间段，也就是确定选上了没有
		if isStable {
			events := make([]event.CourseFromXkEvent, 0, len(courseSubscriptions))
			for _, c := range courseSubscriptions {
				events = append(events, event.CourseFromXkEvent{
					CourseId: c.Course.Id,
					Uid:      uid[0],
					Year:     c.Year,
					Term:     c.Term,
				})
			}
			er := f.producer.BatchProduceCourseListEvent(ctx, events)
			if er != nil {
				f.l.Error("生产CourseListEvent失败", logger.Error(err), logger.String("studentId", studentId))
			}
		}
	case ccnuv1.IsNetworkToXkError(err):
		// 降级,从数据查旧的数据，没查到就直接返回
		var er error
		courseSubscriptions, er = f.CourseService.FindSubscriptionsByUidYearTermAlive(ctx, uid[0], year, term, -1)
		if er != nil {
			return nil, er
		}
		if len(courseSubscriptions) == 0 {
			return nil, ErrDownGradeCourseNotFound
		}
	}
	return courseSubscriptions, err
}
