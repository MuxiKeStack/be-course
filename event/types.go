package event

type CourseFromXkEvent struct {
	CourseId int64
	Uid      int64
	Year     string
	Term     string
}

func (e *CourseFromXkEvent) Topic() string {
	return "course_list_events"
}
