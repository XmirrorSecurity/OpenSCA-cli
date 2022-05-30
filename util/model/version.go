/*
 * @Descripation: 版本号
 * @Date: 2021-11-03 16:03:06
 */

package model

import (
	"strconv"
	"strings"
)

// Version 组件依赖版本号
type Version struct {
	Org    string `json:"org"`
	Nums   []int  `json:"nums,omitempty"`
	Suffix string `json:"suffix,omitempty"`
}

// weight 获取当前版本的后缀权重
func (ver *Version) weight() (weight int) {
	if len(ver.Suffix) > 0 {
		// 后缀权重
		suffixs := map[string]int{"alpha": 1, "beta": 2, "milestone": 3, "rc": 4, "cr": 4, "snapshot": 5, "release": 6, "final": 6, "ga": 6, "sp": 7}
		if w, ok := suffixs[ver.Suffix]; ok {
			// 后缀在后缀列表中取对应后缀权重
			weight = w
		} else {
			// 后缀不在后缀列表中
			weight = 8
		}
	} else {
		// 不存在后缀
		weight = 6
	}
	return weight
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
	length := len(ver.Nums)
	if length > len(other.Nums) {
		length = len(other.Nums)
	}
	// 比较数字大小
	for i := 0; i < length; i++ {
		if ver.Nums[i] < other.Nums[i] {
			return true
		} else if ver.Nums[i] > other.Nums[i] {
			return false
		}
	}
	// 数字多时查看是否有非零值
	if len(ver.Nums) < len(other.Nums) {
		for i := len(other.Nums) - 1; i >= len(ver.Nums); i-- {
			if other.Nums[i] != 0 {
				return true
			}
		}
	}
	// 比较后缀
	vw, ow := ver.weight(), other.weight()
	if vw == ow {
		return ver.Suffix < other.Suffix
	} else {
		return vw < ow
	}
}

// Equal 判断是否等于另一个版本号
func (ver *Version) Equal(other *Version) bool {
	if len(ver.Nums) != len(other.Nums) {
		return false
	}
	// 比较数字大小
	for i, n := range ver.Nums {
		if other.Nums[i] != n {
			return false
		}
	}
	// 比较后缀
	vw, ow := ver.weight(), other.weight()
	return vw == ow
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
