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
	SubscriptionList(ctx context.Context, studentId string, password string, year string,
		term string, uid ...int64) ([]domain.CourseSubscription, error)
	GetDetailById(ctx context.Context, id int64) (domain.Course, error) //在这里面包括成绩
	FindIdOrCreateByCourse(ctx context.Context, course domain.Course) (int64, error)
	FindIdOrUpsertByCourse(ctx context.Context, course domain.Course) (int64, error)
	GetSubscriberUidsByCourseId(ctx context.Context, courseId int64, curUid int64, limit int64) ([]int64, error)
	// FindSubscriptionsByUidYearTermAlive TTL 为-1表示永不过期
	FindSubscriptionsByUidYearTermAlive(ctx context.Context, uid int64, year string, term string,
		TTL time.Duration) ([]domain.CourseSubscription, error)
	Subscribed(ctx context.Context, uid int64, courseId int64) (bool, error)
}

type courseService struct {
	ccnu        ccnuv1.CCNUServiceClient
	repo        repository.CourseRepository
	subRepo     repository.CourseSubscriptionRepository
	currentYear string
	currentTerm string
}

func (s *courseService) Subscribed(ctx context.Context, uid int64, courseId int64) (bool, error) {
	return s.subRepo.Subscribed(ctx, uid, courseId)
}

func NewCourseService(ccnu ccnuv1.CCNUServiceClient, repo repository.CourseRepository, subRepo repository.CourseSubscriptionRepository,
	currentYear string, currentTerm string) CourseService {
	return &courseService{ccnu: ccnu, repo: repo, subRepo: subRepo, currentYear: currentYear, currentTerm: currentTerm}
}

func (s *courseService) SubscriptionList(ctx context.Context, studentId string, password string, year string,
	term string, uid ...int64) ([]domain.CourseSubscription, error) {
	// 从课程接口，判断是否选课中，好像都无所谓，都返回就行了，但是后面发表课评的时候要判断是否选课中
	isHistory := year < s.currentYear || year == s.currentYear && term < s.currentTerm || year == "" || term == ""
	var src ccnuv1.Source
	// 判断学年期，从成绩接口还是老接口
	if isHistory {
		src = ccnuv1.Source_GradeApi
	} else {
		src = ccnuv1.Source_OldXkApi
	}
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
	courseSubscriptions := slice.Map(res.Courses, func(idx int, src *ccnuv1.Course) domain.CourseSubscription {
		return domain.CourseSubscription{
			Course: domain.Course{
				CourseCode: src.CourseCode,
				Name:       src.Name,
				Teacher:    src.Teacher,
				School:     src.School,
				Property:   domain.CoursePropertyFromStr(src.Property),
				Credit:     src.Credit,
			},
			//Uid: uid[0],    // 这个不一定需要因为调用方一定知道自己的uid
			Year: src.Year,
			Term: src.Term,
		}
	})

	// 要在这里聚合出courseId，两种查询结果要采用不同的聚合手段,两个不同的聚合id的接口	[优胜劣汰]
	var eg errgroup.Group
	if isHistory {
		for i := range courseSubscriptions {
			eg.Go(func() error {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
				defer cancel()
				courseSubscriptions[i].Course.Id, err = s.FindIdOrUpsertByCourse(ctx, courseSubscriptions[i].Course)
				return err
			})
		}
	} else {
		for i := range courseSubscriptions {
			eg.Go(func() error {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
				defer cancel()
				courseSubscriptions[i].Course.Id, err = s.FindIdOrCreateByCourse(ctx, courseSubscriptions[i].Course)
				return err
			})
		}
	}
	err = eg.Wait()
	if err != nil {
		return nil, err
	}
	return courseSubscriptions, err
}

func (s *courseService) GetDetailById(ctx context.Context, id int64) (domain.Course, error) {
	return s.repo.FindById(ctx, id)
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

// FindByUidYearTermAlive year , term , TTL 可以为空，代表全部
func (s *courseService) FindSubscriptionsByUidYearTermAlive(ctx context.Context, uid int64, year string, term string,
	TTL time.Duration) ([]domain.CourseSubscription, error) {
	// 不使用join语句，分步：先拿到courseIds
	courseSubs, err := s.subRepo.FindByUidYearTermAlive(ctx, uid, year, term, TTL)
	if err != nil {
		return nil, err
	}
	// 填充具体的course信息
	for i := range courseSubs {
		var er error
		courseSubs[i].Course, er = s.repo.FindById(ctx, courseSubs[i].Course.Id)
		if er != nil {
			return nil, err
		}
	}
	return courseSubs, nil
}

func (s *courseService) GetSubscriberUidsByCourseId(ctx context.Context, courseId int64, curUid int64, limit int64) ([]int64, error) {
	return s.subRepo.FindSubscriberUidsByCourseId(ctx, courseId, curUid, limit)
}
