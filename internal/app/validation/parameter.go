package validation

import (
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

func ParameterToBool(inputValue string) bool {
	if inputValue == "true" || inputValue == "yes" {
		return true
	}
	return false
}

func ParseParamertToStr(inpuStr string) string {
	inpuStr = strings.Trim(inpuStr, "=")
	inpuStr = strings.ReplaceAll(inpuStr, "==", "=")
	arr := strings.Split(inpuStr, "=")
	if len(arr) > 1 {
		return arr[1]
	}
	return ""
}

func ParseParamertToComment(inpuStr string) string {
	inpuStr = strings.Trim(inpuStr, "=")
	inpuStr = strings.ReplaceAll(inpuStr, "==", "=")
	arr := strings.Split(inpuStr, "=")
	if len(arr) > 1 {
		return strings.Join(arr[1:], "=")
	}
	return ""
}

func ParseParamertToUint(inputValue string) uint64 {
	var q uint64
	inputValue = strings.Trim(inputValue, "=")
	inputValue = strings.ReplaceAll(inputValue, "==", "=")
	Arr := strings.Split(inputValue, "=")
	if len(Arr) > 1 {
		quotaStr := Arr[1]
		q = ParamertToUint(quotaStr)
		return q
	}
	return q
}

func ParamertToUint(inputValue string) uint64 {
	inputValue = strings.Trim(inputValue, "\r")
	quota, err := strconv.ParseUint(inputValue, 10, 64)
	if err != nil {
		quotaF, err2 := strconv.ParseFloat(inputValue, 64)
		if err != nil {
			logrus.Errorf("Error parse quota from input string(%v):(%v)(%v)", inputValue, err, err2)
			return 0
		}
		quota = uint64(quotaF)
	}
	return quota
}

func BoolToParamert(trigger bool) string {
	if trigger {
		return "yes"
	}
	return "no"
}
