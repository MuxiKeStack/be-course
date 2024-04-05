package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type CourseDAO interface {
	BatchInsert(ctx context.Context, courses []FailOverCourse) error
	FindByStudentIdYearTerm(ctx context.Context, studentId string, year string, term string) ([]FailOverCourse, error)
}

type GORMCourseDAO struct {
	db *gorm.DB
}

func (dao *GORMCourseDAO) BatchInsert(ctx context.Context, courses []FailOverCourse) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		for _, c := range courses {
			// upsert 语义，其实这里世界create也可以，但是可能未来会对更新时间有要求
			// 比如非课程变动期判断utime足够新就不去ccnu.xk查询了
			err = tx.Clauses(
				clause.OnConflict{DoUpdates: clause.Assignments(map[string]any{
					"utime": now,
				})},
			).Create(&c).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (dao *GORMCourseDAO) FindByStudentIdYearTerm(ctx context.Context, studentId string, year string, term string) ([]FailOverCourse, error) {
	var fcs []FailOverCourse
	err := dao.db.WithContext(ctx).
		Where("student_id = ? and year = ? and term = ?", studentId, year, term).
		Find(&fcs).Error
	return fcs, err
}

type FailOverCourse struct {
	Id        int64  `gorm:"primaryKey,autoIncrement"`
	StudentId string `gorm:"index:sid_year_term;uniqueIndex:sid_cid_name_teacher"`
	Year      string `gorm:"index:sid_year_term"` // 冗余这两个字段便于查询
	Term      string `gorm:"index:sid_year_term"` //
	CourseId  string `gorm:"uniqueIndex:sid_cid_name_teacher"`
	Name      string `gorm:"uniqueIndex:sid_cid_name_teacher"`
	Teacher   string `gorm:"uniqueIndex:sid_cid_name_teacher"`
	Ctime     int64
	Utime     int64
}
