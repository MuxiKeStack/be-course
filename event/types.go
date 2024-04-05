package event

import "github.com/MuxiKeStack/be-course/domain"

type CourseListEvent struct {
	Courses []domain.FailOverCourse
}

func (e *CourseListEvent) Topic() string {
	return "course_list_events"
}
