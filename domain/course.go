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

// TODO 接下来在这里做好类型的转换，外部不需要修改
func (p CourseProperty) String() string {
	return ""
}

func (p CourseProperty) Uint8() uint8 {
	return 0
}

const (
	CoursePropertyUnknown CourseProperty = 0 + iota
)

// CoursePropertyFromStr 将外部调用(ccnu调用)获取到的字符串课程转换为uint8
func CoursePropertyFromStr(pStr string) CourseProperty {
	return CoursePropertyUnknown
	// TODO 课程性质转换
	//switch pStr {
	//case "通识核心课":
	//	return
	//case :
	//
	//case "通识必修课":
	//case "专业主干课":
	//case "个性发发展课":
	//case "通识核心选修课":
	//}
}
