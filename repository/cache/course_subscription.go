package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

var ErrKeyNotExist = redis.Nil

type CourseSubscriptionCache interface {
	GetFirstPageUids(ctx context.Context, courseId int64) ([]int64, error)
	SetFirstPageUids(ctx context.Context, courseId int64, uids []int64) error
}

type RedisCourseSubscriptionCache struct {
	cmd redis.Cmdable
}

func NewRedisCourseSubscriptionCache(cmd redis.Cmdable) CourseSubscriptionCache {
	return &RedisCourseSubscriptionCache{cmd: cmd}
}

func (cache *RedisCourseSubscriptionCache) GetFirstPageUids(ctx context.Context, courseId int64) ([]int64, error) {
	key := cache.firstPageUidsKey(courseId)
	val, err := cache.cmd.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}
	var uids []int64
	err = json.Unmarshal(val, &uids)
	return uids, err
}

func (cache *RedisCourseSubscriptionCache) SetFirstPageUids(ctx context.Context, courseId int64, uids []int64) error {
	key := cache.firstPageUidsKey(courseId)
	val, err := json.Marshal(uids)
	if err != nil {
		return err
	}
	return cache.cmd.Set(ctx, key, val, time.Minute*1).Err() // 半天过期
}

func (cache *RedisCourseSubscriptionCache) firstPageUidsKey(courseId int64) string {
	return fmt.Sprintf("kstack:courses:%d:first_page_uids", courseId)
}
