package domain

import coursev1 "github.com/MuxiKeStack/be-api/gen/proto/course/v1"

type CourseSubscription struct {
	Course Course
	Uid    int64
	Year   string
	Term   string
}

type Course struct {
	Id         int64
	CourseCode string
	Name       string
	Teacher    string
	School     string
	Property   coursev1.CourseProperty
	Credit     float64
}

// CoursePropertyFromStr 将外部调用(ccnu调用)获取到的字符串课程转换为enum CourseProperty
func CoursePropertyFromStr(pStr string) coursev1.CourseProperty {
	switch pStr {
	case "通识核心课":
		return coursev1.CourseProperty_CoursePropertyGeneralCore
	case "通识选修课":
		return coursev1.CourseProperty_CoursePropertyGeneralElective
	case "通识必修课":
		return coursev1.CourseProperty_CoursePropertyGeneralRequired
	case "专业主干课程":
		return coursev1.CourseProperty_CoursePropertyMajorCore
	case "个性发展课程":
		return coursev1.CourseProperty_CoursePropertyMajorElective
	default:
		return coursev1.CourseProperty_CoursePropertyUnknown
	}
}
