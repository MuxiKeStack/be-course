package dao

import (
	"context"
	"errors"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

var (
	ErrRecordNorFound  = gorm.ErrRecordNotFound
	ErrCourseDuplicate = errors.New("课程创建冲突")
)

type CourseDAO interface {
	FindGradesById(ctx context.Context, id int64) ([]Grade, error)
	FindById(ctx context.Context, id int64) (Course, error)
	FindByIds(ctx context.Context, cids []int64) ([]Course, error)
	FindIdByCourse(ctx context.Context, course Course) (int64, error)
	Insert(ctx context.Context, course Course) error
	// BatchUpsert 这个实际上并未被上层的repository使用，而是被导入课程的脚本所使用的
	// 理论上每一个数据库只能由其微服务来调用，不能跨过服务直接调其数据库
	// 但这里调用方是一个本地手动执行的脚本不是一个微服务，从实用性和效率上讲就这样写了
	BatchUpsert(ctx context.Context, courses []Course) error
	Upsert(ctx context.Context, course Course) error
	FindIdByCourseWithoutUnknownProperty(ctx context.Context, course Course) (int64, error)
}

type GORMCourseDAO struct {
	db *gorm.DB
}

func NewGORMCourseDAO(db *gorm.DB) CourseDAO {
	return &GORMCourseDAO{db: db}
}

func (dao *GORMCourseDAO) Upsert(ctx context.Context, course Course) error {
	now := time.Now().UnixMilli()
	course.Utime = now
	course.Ctime = now
	return dao.db.WithContext(ctx).Clauses(
		clause.OnConflict{DoUpdates: clause.Assignments(map[string]any{
			"property": course.Property,
			"utime":    now,
		})}).Create(&course).Error
}

func (dao *GORMCourseDAO) FindIdByCourse(ctx context.Context, course Course) (int64, error) {
	var id int64
	err := dao.db.WithContext(ctx).
		Model(&Course{}).
		Select("id").
		Where("course_code = ? and name = ? and teacher = ?", course.CourseCode, course.Name, course.Teacher).
		First(&id).Error
	return id, err
}

func (dao *GORMCourseDAO) FindIdByCourseWithoutUnknownProperty(ctx context.Context, course Course) (int64, error) {
	var id int64
	const CoursePropertyUnknown = 0
	err := dao.db.WithContext(ctx).
		Model(&Course{}).
		Select("id").
		Where("course_code = ? and name = ? and teacher = ? and property != ?",
			course.CourseCode, course.Name, course.Teacher, CoursePropertyUnknown).
		First(&id).Error
	return id, err
}

func (dao *GORMCourseDAO) Insert(ctx context.Context, course Course) error {
	now := time.Now().UnixMilli()
	course.Ctime = now
	course.Utime = now
	return dao.db.WithContext(ctx).Create(&course).Error
}

func (dao *GORMCourseDAO) FindByIds(ctx context.Context, cids []int64) ([]Course, error) {
	var courses []Course
	err := dao.db.WithContext(ctx).
		Where("id in ?", cids).
		Find(&courses).Error
	return courses, err
}

func (dao *GORMCourseDAO) FindGradesById(ctx context.Context, id int64) ([]Grade, error) {
	var grades []Grade
	err := dao.db.WithContext(ctx).
		Where("course_id = ?", id).
		Find(&grades).Error
	return grades, err
}

func (dao *GORMCourseDAO) FindById(ctx context.Context, id int64) (Course, error) {
	var c Course
	err := dao.db.WithContext(ctx).
		Where("id = ?", id).
		First(&c).Error
	return c, err
}

func (dao *GORMCourseDAO) BatchUpsert(ctx context.Context, courses []Course) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var eg errgroup.Group
		for _, c := range courses {
			c.Ctime = now
			c.Utime = now
			eg.Go(func() error {
				return tx.Clauses(clause.OnConflict{DoUpdates: clause.Assignments(map[string]any{
					"utime": now,
				})}).Create(&c).Error
			})
		}
		return eg.Wait()
	})
}

type Course struct {
	Id int64 `gorm:"primaryKey,autoIncrement"`
	// 这里是否有必要为property建立一个包含四个字段的联合索引
	CourseCode string `gorm:"uniqueIndex:courseCode_name_teacher; index:idx_code_name_teacher_property; type:char(8)"`
	Name       string `gorm:"uniqueIndex:courseCode_name_teacher; index:idx_code_name_teacher_property; type:varchar(30)"`
	Teacher    string `gorm:"uniqueIndex:courseCode_name_teacher; index:idx_code_name_teacher_property; type:varchar(10)"`
	Property   int32  `gorm:"index:idx_code_name_teacher_property"`
	School     string
	Credit     float32
	Ctime      int64
	Utime      int64
}

type Grade struct {
	Id       int64 `gorm:"primaryKey,autoIncrement"`
	CourseId int64 `gorm:"uniqueIndex:cid_uid;"` // 主键id
	Uid      int64 `gorm:"uniqueIndex:cid_uid"`
	Regular  float32
	Final    float32
	Total    float32
	Year     string
	Term     string
	Utime    int64
	Ctime    int64
}
