package validation

import (
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// 自定义验证器正则表达式
var (
	mobileCnRegex = regexp.MustCompile(`^1[3-9]\d{9}$`) // 中国大陆手机号正则表达式
)

// Init 初始化验证器
func Init() error {
	// 获取验证器实例
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return nil
	}

	// 注册手机号验证
	return v.RegisterValidation("mobile_cn", validateMobileCn)
}

// validateMobileCn 验证中国大陆手机号
func validateMobileCn(fl validator.FieldLevel) bool {
	return mobileCnRegex.MatchString(fl.Field().String())
}
