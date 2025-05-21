// Package validation 提供自定义验证器功能，扩展Gin框架的数据验证能力
package validation

import (
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// 自定义验证器正则表达式
var (
	// 中国大陆手机号正则表达式
	mobileCnRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)
)

// Init 初始化验证器，注册自定义验证规则
func Init() error {
	// 获取验证器引擎
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return nil
	}

	// 注册手机号验证规则 `binding:"mobile_cn"`
	return v.RegisterValidation("mobile_cn", validateMobileCn)
}

// validateMobileCn 验证中国大陆手机号格式
func validateMobileCn(fl validator.FieldLevel) bool {
	return mobileCnRegex.MatchString(fl.Field().String())
}
