package service

import (
	"errors"
	"fmt"
	"time"

	"app/config"
	"app/internal/constant"
	"app/internal/dto"
	"app/internal/model"
	"app/internal/repository"
	"app/internal/utils"
	"app/pkg/jwt"
	"app/pkg/logger"
	"app/pkg/redis"
	"app/pkg/sms"
)

var (
	// ErrUserNotFound 用户不存在错误
	ErrUserNotFound = errors.New(constant.ErrUserNotFound)
	// ErrInvalidCode 验证码无效错误
	ErrInvalidCode = errors.New(constant.ErrInvalidCode)
	// ErrDeactivateFailed 注销失败错误
	ErrDeactivateFailed = errors.New(constant.ErrDeactivateFailed)
)

// TokenBlacklistPrefix 令牌黑名单前缀，使用常量包中的定义
const TokenBlacklistPrefix = constant.TokenBlacklistPrefix

// UserService 用户服务接口
type UserService interface {
	// SendVerificationCode 发送验证码
	SendVerificationCode(req *dto.SendVerificationCodeRequest) (*dto.SendVerificationCodeResponse, error)
	// VerificationCodeLogin 验证码登录
	VerificationCodeLogin(req *dto.VerificationCodeLoginRequest) (*dto.LoginResponse, error)
	// GetUserInfo 获取用户信息
	GetUserInfo(id uint) (*dto.UserInfoResponse, error)
	// DeactivateAccount 注销账号
	DeactivateAccount(req *dto.DeactivateAccountRequest) error
	// Logout 退出登录
	Logout(req *dto.LogoutRequest) (*dto.LogoutResponse, error)
}

// userService 用户服务实现
type userService struct {
	userRepo repository.UserRepository
	smsRepo  repository.SMSRepository
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repository.UserRepository, smsRepo repository.SMSRepository) UserService {
	return &userService{
		userRepo: userRepo,
		smsRepo:  smsRepo,
	}
}

// SendVerificationCode 发送验证码
func (s *userService) SendVerificationCode(req *dto.SendVerificationCodeRequest) (*dto.SendVerificationCodeResponse, error) {
	logger.WithField("mobile", req.Mobile).WithField("type", req.Type).Info("开始发送验证码")

	// 生成随机验证码
	code := generateVerificationCode(constant.VerificationCodeLength)

	// 根据验证码类型选择不同的前缀
	var prefix string
	switch req.Type {
	case dto.VerificationTypeLogin:
		prefix = constant.VerificationCodePrefixLogin
	case dto.VerificationTypeDeactivate:
		prefix = constant.VerificationCodePrefixDeactivate
	default:
		prefix = constant.VerificationCodePrefixLogin
	}

	// 将验证码保存到Redis，设置过期时间
	key := prefix + req.Mobile
	err := redis.Set(key, code, constant.VerificationCodeExpiration)
	if err != nil {
		logger.WithError(err).Error("保存验证码到Redis失败")
		return nil, fmt.Errorf("保存验证码失败: %w", err)
	}

	// 发送短信验证码
	client, err := sms.GetSMSClient()
	if err != nil {
		logger.WithError(err).Error("创建短信客户端失败")
		return nil, fmt.Errorf("创建短信客户端失败: %w", err)
	}

	// 从配置中获取验证码短信模板代码
	smsConfig := config.GetSMSConfig()
	templateCode := smsConfig.Aliyun.Templates["verification_code"]
	if templateCode == "" {
		logger.Error("验证码短信模板未配置")
		return nil, fmt.Errorf("短信模板配置错误: %w", err)
	}

	// 根据验证码类型构建不同的短信内容
	var smsContent string
	switch req.Type {
	case dto.VerificationTypeLogin:
		smsContent = fmt.Sprintf("您的登录验证码是：%s，5分钟内有效。", code)
	case dto.VerificationTypeDeactivate:
		smsContent = fmt.Sprintf("您的账号注销验证码是：%s，5分钟内有效。请谨慎操作，注销后账号将无法恢复。", code)
	default:
		smsContent = fmt.Sprintf("您的验证码是：%s，5分钟内有效。", code)
	}

	// 构建短信请求
	smsReq := sms.SMSRequest{
		PhoneNumbers: req.Mobile,
		TemplateCode: templateCode,
		TemplateParam: map[string]string{
			"code": code,
		},
	}

	// 发送短信
	smsResp, err := client.SendSMS(smsReq)
	if err != nil {
		logger.WithError(err).Error("发送短信失败")
		return nil, fmt.Errorf("发送短信失败: %w", err)
	}

	// 记录短信发送信息
	smsRecord := &model.SMSRecord{
		PhoneNumber:   req.Mobile,
		Type:          constant.SMSTypeVerification,
		Content:       smsContent,
		TemplateCode:  templateCode,
		TemplateParam: fmt.Sprintf(`{"code":"%s"}`, code),
		Status:        "success",
		RequestId:     smsResp.RequestId,
		BizId:         smsResp.BizId,
	}

	// 保存短信记录
	err = s.smsRepo.Create(smsRecord)
	if err != nil {
		// 记录失败不影响主流程，只记录错误日志
		logger.WithError(err).Warn("记录短信发送信息失败")
	}

	return &dto.SendVerificationCodeResponse{
		Message: "验证码已发送",
	}, nil
}

// VerificationCodeLogin 验证码登录
func (s *userService) VerificationCodeLogin(req *dto.VerificationCodeLoginRequest) (*dto.LoginResponse, error) {
	logger.WithField("mobile", req.Mobile).Info("验证码登录")

	// 从Redis获取验证码（登录验证码）
	key := constant.VerificationCodePrefixLogin + req.Mobile
	savedCode, err := redis.Get(key)
	if err != nil || savedCode != req.Code {
		logger.WithFields(map[string]interface{}{
			"mobile": req.Mobile,
			"error":  err,
		}).Warn("验证码验证失败")
		return nil, ErrInvalidCode
	}

	// 验证成功后删除验证码
	_, _ = redis.Del(key)

	// 查找用户
	user, err := s.userRepo.FindByMobile(req.Mobile)
	if err != nil {
		logger.WithField("mobile", req.Mobile).Info("用户不存在，创建新用户")
		// 如果用户不存在，则创建新用户
		user = &model.User{
			Mobile:   req.Mobile,
			Username: req.Mobile,                            // 默认使用手机号作为用户名
			Nickname: "用户" + req.Mobile[len(req.Mobile)-4:], // 使用手机号后4位作为昵称
			Status:   constant.UserStatusNormal,             // 正常状态
		}

		// 保存新用户
		err = s.userRepo.Create(user)
		if err != nil {
			logger.WithError(err).Error("创建用户失败")
			return nil, fmt.Errorf("创建用户失败: %w", err)
		}
	}

	// 检查用户状态
	if user.Status != constant.UserStatusNormal {
		logger.WithField("user_id", user.ID).Warn("账号已被禁用")
		return nil, errors.New("账号已被禁用")
	}

	// 生成JWT令牌
	token, err := jwt.GenerateToken(user.ID, user.Username, "")
	if err != nil {
		logger.WithError(err).Error("生成令牌失败")
		return nil, fmt.Errorf("生成令牌失败: %w", err)
	}

	// 构建响应
	response := &dto.LoginResponse{
		Token: token,
	}

	// 填充用户信息
	response.User.ID = user.ID
	response.User.Username = user.Username
	response.User.Mobile = user.Mobile
	response.User.Nickname = user.Nickname
	response.User.Avatar = user.Avatar

	logger.WithField("user_id", user.ID).Info("用户登录成功")
	return response, nil
}

// generateVerificationCode 生成指定长度的随机验证码
func generateVerificationCode(length int) string {
	// 使用utils包中的函数生成随机数字
	return utils.GenerateRandomDigits(length)
}

// GetUserInfo 获取用户信息
func (s *userService) GetUserInfo(id uint) (*dto.UserInfoResponse, error) {
	logger.WithField("user_id", id).Info("获取用户信息")

	// 根据ID查找用户
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			logger.WithField("user_id", id).Warn("用户不存在")
			return nil, ErrUserNotFound
		}
		logger.WithError(err).Error("查询用户失败")
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}

	// 构建响应
	response := &dto.UserInfoResponse{
		ID:        user.ID,
		Username:  user.Username,
		Mobile:    user.Mobile,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Status:    user.Status,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	return response, nil
}

// Logout 退出登录
func (s *userService) Logout(req *dto.LogoutRequest) (*dto.LogoutResponse, error) {
	logger.WithField("user_id", req.UserID).Info("用户退出登录")

	// 解析令牌，获取过期时间
	claims, err := jwt.ParseToken(req.Token)
	if err != nil {
		// 如果令牌已经无效，则直接返回成功
		if err == jwt.ErrTokenInvalid || err == jwt.ErrTokenExpired {
			return &dto.LogoutResponse{Message: "退出登录成功"}, nil
		}
		return nil, fmt.Errorf("解析令牌失败: %w", err)
	}

	// 计算令牌剩余有效期
	expTime := claims.ExpiresAt.Time
	ttl := time.Until(expTime)
	if ttl <= 0 {
		// 令牌已过期，直接返回成功
		return &dto.LogoutResponse{Message: "退出登录成功"}, nil
	}

	// 将令牌加入黑名单，过期时间与令牌相同
	blacklistKey := TokenBlacklistPrefix + req.Token
	err = redis.Set(blacklistKey, "revoked", ttl)
	if err != nil {
		logger.WithError(err).Error("将令牌加入黑名单失败")
		return nil, fmt.Errorf("退出登录失败: %w", err)
	}

	return &dto.LogoutResponse{Message: "退出登录成功"}, nil
}

// DeactivateAccount 注销账号
func (s *userService) DeactivateAccount(req *dto.DeactivateAccountRequest) error {
	logger.WithField("user_id", req.UserID).Info("开始注销账号")

	// 验证验证码（注销验证码）
	key := constant.VerificationCodePrefixDeactivate + req.Mobile
	savedCode, err := redis.Get(key)
	if err != nil || savedCode != req.Code {
		logger.WithFields(map[string]interface{}{
			"mobile": req.Mobile,
			"error":  err,
		}).Warn("验证码验证失败")
		return ErrInvalidCode
	}

	// 验证成功后删除验证码
	_, _ = redis.Del(key)

	// 查找用户
	user, err := s.userRepo.FindByID(req.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			logger.WithField("user_id", req.UserID).Warn("用户不存在")
			return ErrUserNotFound
		}
		logger.WithError(err).Error("查询用户失败")
		return fmt.Errorf("查询用户失败: %w", err)
	}

	// 验证手机号是否匹配
	if user.Mobile != req.Mobile {
		logger.WithFields(map[string]interface{}{
			"user_id":      req.UserID,
			"user_mobile":  user.Mobile,
			"input_mobile": req.Mobile,
		}).Warn("手机号不匹配，注销失败")
		return errors.New("手机号不匹配，注销失败")
	}

	// 执行注销操作（软删除）
	err = s.userRepo.SoftDelete(req.UserID)
	if err != nil {
		logger.WithError(err).Error("注销账号失败")
		return ErrDeactivateFailed
	}

	logger.WithField("user_id", req.UserID).Info("账号注销成功")
	return nil
}
