package event

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/MuxiKeStack/be-course/domain"
	"github.com/MuxiKeStack/be-course/pkg/logger"
	"github.com/MuxiKeStack/be-course/pkg/saramax"
	"github.com/MuxiKeStack/be-course/repository"
	"github.com/ecodeclub/ekit/slice"
	"time"
)

type CourseListEventConsumer struct {
	client sarama.Client
	l      logger.Logger
	repo   repository.CourseSubscriptionRepository
}

// Start 这边就是自己启动 goroutine 了
func (c *CourseListEventConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("store",
		c.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{(&CourseFromXkEvent{}).Topic()},
			saramax.NewBatchHandler(c.l, c.BatchConsume))
		if err != nil {
			c.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (c *CourseListEventConsumer) BatchConsume(msgs []*sarama.ConsumerMessage, events []CourseFromXkEvent) error {
	courseSubscriptions := slice.Map(events, func(c int, src CourseFromXkEvent) domain.CourseSubscription {
		return domain.CourseSubscription{
			CourseId: src.CourseId,
			Uid:      src.Uid,
			Year:     src.Year,
			Term:     src.Term,
		}
	})
	// 批量存储到数据库
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return c.repo.BatchCreateCourseSubscription(ctx, courseSubscriptions)
}
