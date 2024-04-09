package domain

type CourseSubscription struct {
	CourseId int64
	Uid      int64
	Year     string
	Term     string
}

type Course struct {
	Id         int64
	CourseCode string
	Name       string
	Teacher    string
	School     string
	Property   CourseProperty
	Credit     float32
}

type Grade struct {
	Regular float32
	Final   float32
	Total   float32
	Year    string
	Term    string
}

type CourseProperty uint8

const (
	CoursePropertyUnknown         CourseProperty = iota
	CoursePropertyGeneralCore                    // 通识核心课
	CoursePropertyGeneralElective                // 通识选修课
	CoursePropertyGeneralRequired                // 通识必修课
	CoursePropertyMajorCore                      // 专业主干课程
	CoursePropertyMajorElective                  // 个性发展课程
)

func (p CourseProperty) Uint8() uint8 {
	return uint8(p)
}

func (p CourseProperty) String() string {
	switch p {
	case CoursePropertyGeneralCore:
		return "通识核心课"
	case CoursePropertyGeneralElective:
		return "通识选修课"
	case CoursePropertyGeneralRequired:
		return "通识必修课"
	case CoursePropertyMajorCore:
		return "专业主干课程"
	case CoursePropertyMajorElective:
		return "个性发展课程"
	default:
		return "Unknown Course Property"
	}
}

// CoursePropertyFromStr 将外部调用(ccnu调用)获取到的字符串课程转换为uint8
func CoursePropertyFromStr(pStr string) CourseProperty {
	switch pStr {
	case "通识核心课":
		return CoursePropertyGeneralCore
	case "通识选修课":
		return CoursePropertyGeneralElective
	case "通识必修课":
		return CoursePropertyGeneralRequired
	case "专业主干课程":
		return CoursePropertyMajorCore
	case "个性发展课程":
		return CoursePropertyMajorElective
	default:
		return CoursePropertyUnknown
	}
}
