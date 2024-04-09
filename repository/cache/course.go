package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/MuxiKeStack/be-course/domain"
	"github.com/redis/go-redis/v9"
)

type CourseCache interface {
	Get(ctx context.Context, id int64) (domain.Course, error)
	GetGrades(ctx context.Context, cid int64) ([]domain.Grade, error)
}

type RedisCourseCache struct {
	cmd redis.Cmdable
}

func NewRedisCourseCache(cmd redis.Cmdable) CourseCache {
	return &RedisCourseCache{cmd: cmd}
}

func (cache *RedisCourseCache) Get(ctx context.Context, id int64) (domain.Course, error) {
	val, err := cache.cmd.Get(ctx, cache.key(id)).Bytes()
	var c domain.Course
	err = json.Unmarshal(val, &c)
	return c, err
}

func (cache *RedisCourseCache) key(id int64) string {
	return fmt.Sprintf("kstack:courses:%d", id)
}

func (cache *RedisCourseCache) GetGrades(ctx context.Context, cid int64) ([]domain.Grade, error) {
	val, err := cache.cmd.Get(ctx, cache.gradesKey(cid)).Bytes()
	var c []domain.Grade
	err = json.Unmarshal(val, &c)
	return c, err
}

func (cache *RedisCourseCache) gradesKey(cid int64) string {
	return fmt.Sprintf("kstack:courses:grades:%d", cid)
}
