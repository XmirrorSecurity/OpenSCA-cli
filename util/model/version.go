package model

import (
	"regexp"
	"strconv"
	"strings"
)

// Version 组件依赖版本号
type Version struct {
	Org    string `json:"org"`
	Nums   []int  `json:"nums,omitempty"`
	Suffix string `json:"suffix,omitempty"`
}
type token struct {
	// 连接符
	// true 代表 -, false 代表 .
	link bool
	// 值 整数
	num int
	// 值 字符串
	str string
	// 标记是否为值
	isnum bool
}

var (
	// 后缀权重
	suffixs = map[string]int{"alpha": 1, "beta": 2, "milestone": 3, "rc": 4, "cr": 4, "snapshot": 5, "release": 6, "final": 6, "ga": 6, "sp": 7}
	// 数字or字母匹配
	numStrReg = regexp.MustCompile(`((\d+)|([a-zA-Z]+))`)
)

func (t token) compare(t2 token) int {
	// 比较数字
	if t.isnum && !t2.isnum {
		return 1
	} else if !t.isnum && t2.isnum {
		return -1
	} else if t.isnum && t2.isnum {
		if t.num == t2.num {
			if !t.link && t2.link {
				return 1
			} else if t.link && !t2.link {
				return -1
			} else {
				return 0
			}
		} else {
			return t.num - t2.num
		}
	}
	// 比较字符串
	if t.str != t2.str {
		w, ok := suffixs[strings.ToLower(t.str)]
		w2, ok2 := suffixs[strings.ToLower(t2.str)]
		if ok && ok2 {
			return w - w2
		} else if ok && !ok2 {
			return -1
		} else if !ok && ok2 {
			return 1
		}
		if t.str > t2.str {
			return 1
		} else {
			return -1
		}
	}
	// 比较分隔符
	if t.link != t2.link {
		if t.num != 0 {
			// 数字.分隔符优先级高
			if !t.link {
				return 1
			} else {
				return -1
			}
		}
		if t.str != "" {
			// 字符串-分隔符优先级高
			if t.link {
				return 1
			} else {
				return -1
			}
		}
	}
	return 0
}

// compareToken 比较两组token
// return a - b
func compareToken(a, b []token) int {
	var min int
	if len(a) > len(b) {
		if a[len(b)].str != "" {
			b = append(b, token{link: true, str: "ga"})
		}
		min = len(b)
	} else if len(a) < len(b) {
		if b[len(a)].str != "" {
			a = append(a, token{link: true, str: "ga"})
		}
		min = len(a)
	} else {
		min = len(a)
	}
	// 依次比较token
	for i := 0; i < min; i++ {
		r := a[i].compare(b[i])
		if r != 0 {
			return r
		}
	}
	// 返回长的那个
	return len(a) - len(b)
}

// parseToken 从版本号字符串中解析token
func parseToken(ver string) (tokens []token) {
	ver = strings.ToLower(strings.TrimLeft(ver, "vV"))
	tokens = []token{}
	t := token{isnum: true}
	for len(ver) > 0 {
		// 按-和.分割
		index := strings.IndexAny(ver, `.-`)
		for index == 0 {
			next := strings.IndexAny(ver[1:], `.-`)
			if next == -1 {
				index = len(ver)
			} else {
				// 从ver[1:]开始搜索，所以需要下标+1
				index = next + 1
			}
		}
		if index == -1 {
			index = len(ver)
		}
		word := ver[:index]
		ver = ver[index:]
		// 检测到分隔符重新创建新token
		if word[0] == '.' || word[0] == '-' {
			tokens = append(tokens, t)
			t = token{link: word[0] == '-', isnum: word[0] == '.'}
			word = word[1:]
		}
		// 尝试解析数字
		if n, err := strconv.Atoi(word); err == nil {
			t.num = n
			t.isnum = true
		} else if !strings.ContainsAny(word, `1234567890`) {
			// 不含数字则保存限定符
			t.str = word
		} else {
			// 标记下一个token是否是额外创建的'-'分隔符
			link := false
			// 解析数字与字符串
			matchs := numStrReg.FindAllString(word, -1)
			for i, match := range matchs {
				if n, err := strconv.Atoi(match); err == nil {
					t.num = n
					t.isnum = true
				} else {
					// 为单个字母并后面存在数字
					if len(match) == 1 && i+1 < len(matchs) {
						if match == "a" {
							match = "alpha"
						} else if match == "b" {
							match = "beta"
						} else if match == "m" {
							match = "milestone"
						}
					}
					t.str = match
				}
				tokens = append(tokens, t)
				t = token{link: true}
				link = true
			}
			if link {
				t.link = false
			}
		}
	}
	tokens = append(tokens, t)
	// 处理限定符
	for i := range tokens {
		if tokens[i].str != "" {
			s := tokens[i].str
			if s == "final" || s == "ga" {
				s = ""
			}
			tokens[i].str = s
			tokens[i].isnum = false
		}
	}
	isZero := true
	for i := len(tokens) - 1; i >= 0; i-- {
		t := tokens[i]
		if t.num == 0 {
			if t.str == "" {
				if isZero || !t.isnum {
					tokens = append(tokens[:i], tokens[i+1:]...)
				}
			} else if t.str != "" {
				isZero = true
			}
		} else {
			isZero = false
		}
	}
	return
}

// NewVersion 解析版本字符串
func NewVersion(verStr string) *Version {
	verStr = strings.TrimSpace(verStr)
	ver := &Version{Nums: []int{}, Org: verStr}
	verStr = strings.TrimLeft(verStr, "vV^~=<>")
	// 获取后缀
	index := strings.Index(verStr, "-")
	if index != -1 {
		ver.Suffix = verStr[index+1:]
		verStr = verStr[:index]
	}
	// 解析版本号
	tags := strings.Split(verStr, ".")
	for i, numStr := range tags {
		if num, err := strconv.Atoi(numStr); err == nil {
			ver.Nums = append(ver.Nums, num)
		} else {
			ver.Suffix = strings.Join(tags[i:], ".")
			break
		}
	}
	// 去除结尾零值
	for len(ver.Nums) > 1 {
		length := len(ver.Nums)
		if ver.Nums[length-1] == 0 {
			ver.Nums = ver.Nums[:length-1]
		} else {
			break
		}
	}
	return ver
}

// Less 判断是否严格小于另一个版本号
func (ver *Version) Less(other *Version) bool {
	va := strings.TrimLeft(ver.Org, "vV^<>=~!, ")
	vb := strings.TrimLeft(other.Org, "vV^<>=~!, ")
	ta := parseToken(va)
	tb := parseToken(vb)
	return compareToken(ta, tb) < 0
}

// Equal 判断是否等于另一个版本号
func (ver *Version) Equal(other *Version) bool {
	if len(ver.Nums) != len(other.Nums) {
		return false
	}
	va := strings.TrimLeft(ver.Org, "vV^<>=~!, ")
	vb := strings.TrimLeft(other.Org, "vV^<>=~!, ")
	ta := parseToken(va)
	tb := parseToken(vb)
	return compareToken(ta, tb) == 0
}

// InRangeInterval 判断一个版本是否在一个版本区间内
func InRangeInterval(ver *Version, interval string) bool {
	// 当前版本
	// 遍历所有区间
	for _, interval := range strings.Split(interval, "||") {
		if len(interval) < 2 {
			continue
		}
		// 判断左边界是否为闭
		left := interval[0] == '['
		// 判断右边界是否为闭
		right := interval[len(interval)-1] == ']'
		// 逗号所在位置
		index := strings.Index(interval, ",")
		if index == -1 {
			return false
		}
		// 区间左值
		leftValue := NewVersion(interval[1:index])
		// 区间右值
		rightValue := NewVersion(interval[index+1 : len(interval)-1])
		// 判断是否在区间边界
		if (left && ver.Equal(leftValue)) || (right && ver.Equal(rightValue)) {
			return true
		}
		// 判断是否在区间内部
		// 大于左值并(右值为空或小于右值)
		// leftValue < version && ( isempty(rightValue) || version < rightValue )
		if leftValue.Less(ver) && (len(rightValue.Nums) == 0 || ver.Less(rightValue)) {
			return true
		}
	}
	// 不在任何一个区间内则返回false
	return false
}

// Ok 检测是否为合法版本号
func (v *Version) Ok() bool {
	return !strings.Contains(v.Org, "$") && len(v.Nums) > 0
}
