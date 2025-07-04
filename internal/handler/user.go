package handler

import (
	"strconv"
	"strings"

	"app/internal/dto"
	"app/internal/service"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户处理器，负责处理用户相关的HTTP请求
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler 创建用户处理器实例
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// SendVerificationCode 发送验证码
func (h *UserHandler) SendVerificationCode(c *gin.Context) {
	var req dto.SendVerificationCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误", err)
		return
	}

	// 发送验证码
	resp, err := h.userService.SendVerificationCode(c, &req)
	if err != nil {
		response.InternalServerError(c, "发送验证码失败", err)
		return
	}

	response.Success(c, resp.Message, nil)
}

// VerificationCodeLogin 验证码登录
func (h *UserHandler) VerificationCodeLogin(c *gin.Context) {
	var req dto.VerificationCodeLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误", err)
		return
	}

	// 验证码登录
	resp, err := h.userService.VerificationCodeLogin(c, &req)
	if err != nil {
		// 根据错误类型设置不同的状态码和错误消息
		switch err {
		case service.ErrInvalidCode:
			response.BadRequest(c, "验证码无效或已过期", err)
		case service.ErrUserNotFound:
			response.NotFound(c, "用户不存在", err)
		default:
			response.InternalServerError(c, "登录失败", err)
		}
		return
	}

	response.Success(c, "登录成功", resp)
}

// Logout 退出登录
func (h *UserHandler) Logout(c *gin.Context) {
	var req dto.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误", err)
		return
	}

	// 从上下文中获取当前用户ID
	currentUserID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "未授权访问", nil)
		return
	}

	// 权限检查：用户只能退出自己的登录
	if currentUserID.(uint) != req.UserID {
		response.Forbidden(c, "权限不足，无法退出其他用户的登录", nil)
		return
	}

	// 获取请求头中的令牌
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response.BadRequest(c, "未提供授权令牌", nil)
		return
	}

	// 提取令牌
	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		response.BadRequest(c, "无效的授权格式", nil)
		return
	}

	// 设置令牌到请求中
	req.Token = parts[1]

	// 退出登录
	resp, err := h.userService.Logout(c, &req)
	if err != nil {
		response.InternalServerError(c, "退出登录失败", err)
		return
	}

	response.Success(c, resp.Message, nil)
}

// DeactivateAccount 注销账号
func (h *UserHandler) DeactivateAccount(c *gin.Context) {
	var req dto.DeactivateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误", err)
		return
	}

	// 从上下文中获取当前用户ID
	currentUserID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "未授权访问", nil)
		return
	}

	// 权限检查：用户只能注销自己的账号
	if currentUserID.(uint) != req.UserID {
		response.Forbidden(c, "权限不足，无法注销其他用户账号", nil)
		return
	}

	// 注销账号
	err := h.userService.DeactivateAccount(c, &req)
	if err != nil {
		// 根据错误类型设置不同的状态码和错误消息
		switch err {
		case service.ErrInvalidCode:
			response.BadRequest(c, "验证码无效或已过期", err)
		case service.ErrUserNotFound:
			response.NotFound(c, "用户不存在", err)
		case service.ErrDeactivateFailed:
			response.InternalServerError(c, "注销账号失败", err)
		default:
			response.InternalServerError(c, "注销账号失败", err)
		}
		return
	}

	response.Success(c, "账号已成功注销", nil)
}

// GetUserInfo 获取用户信息，仅允许用户查看自己的信息
func (h *UserHandler) GetUserInfo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的用户ID", err)
		return
	}

	currentUserID, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "未授权访问", nil)
		return
	}

	if currentUserID.(uint) != uint(id) {
		response.Forbidden(c, "权限不足，无法查看其他用户信息", nil)
		return
	}

	resp, err := h.userService.GetUserInfo(c, uint(id))
	if err != nil {
		if err == service.ErrUserNotFound {
			response.NotFound(c, "用户不存在", err)
		} else {
			response.InternalServerError(c, "获取用户信息失败", err)
		}
		return
	}

	response.Success(c, "获取用户信息成功", resp)
}
