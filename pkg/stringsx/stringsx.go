package stringsx

import "unicode"

// ContainsDigit 检查字符串 s 是否包含至少一个数字。
func ContainsDigit(s string) bool {
	for _, char := range s {
		if unicode.IsDigit(char) {
			return true
		}
	}
	return false
}
