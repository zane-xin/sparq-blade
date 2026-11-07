// Package literal 保存编译和运行阶段共同使用的字面量语法。
package literal

import (
	"strconv"
	"strings"
)

// ParseString 解析单引号或双引号字符串，保留现有无法反转义时返回原始内容的语义。
func ParseString(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if len(value) < 2 {
		return "", false
	}

	if (value[0] == '\'' && value[len(value)-1] == '\'') || (value[0] == '"' && value[len(value)-1] == '"') {
		unquoted, err := strconv.Unquote(`"` + strings.ReplaceAll(value[1:len(value)-1], `"`, `\"`) + `"`)
		if err != nil {
			return value[1 : len(value)-1], true
		}
		return unquoted, true
	}

	return "", false
}
