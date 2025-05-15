// Package validation 提供自定义验证器功能，扩展Gin框架的数据验证能力
package validation

import (
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// 自定义验证器正则表达式
var (
	// mobileCnRegex 中国大陆手机号正则表达式
	// 匹配1开头，第二位为3-9，后面跟9位数字的手机号
	mobileCnRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)
)

// Init 初始化验证器，注册自定义验证规则
// 返回可能的错误，如果验证器引擎获取失败则返回nil
func Init() error {
	// 获取Gin框架的验证器引擎
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		// 如果类型断言失败，说明验证器引擎不是预期的类型
		// 这种情况通常不会发生，除非Gin框架更改了验证器实现
		return nil
	}

	// 注册中国大陆手机号验证规则
	// 使用方式: `binding:"mobile_cn"`
	return v.RegisterValidation("mobile_cn", validateMobileCn)
}

// validateMobileCn 验证中国大陆手机号格式是否正确
// 参数:
//   - fl: validator库的字段级别上下文，包含要验证的字段值
//
// 返回:
//   - 布尔值，表示验证是否通过
func validateMobileCn(fl validator.FieldLevel) bool {
	// 获取字段值并转为字符串，然后用正则表达式匹配
	return mobileCnRegex.MatchString(fl.Field().String())
}
