package service

import (
	"context"
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
	SendVerificationCode(ctx context.Context, req *dto.SendVerificationCodeRequest) (*dto.SendVerificationCodeResponse, error)
	// VerificationCodeLogin 验证码登录
	VerificationCodeLogin(ctx context.Context, req *dto.VerificationCodeLoginRequest) (*dto.LoginResponse, error)
	// GetUserInfo 获取用户信息
	GetUserInfo(ctx context.Context, id uint) (*dto.UserInfoResponse, error)
	// DeactivateAccount 注销账号
	DeactivateAccount(ctx context.Context, req *dto.DeactivateAccountRequest) error
	// Logout 退出登录
	Logout(ctx context.Context, req *dto.LogoutRequest) (*dto.LogoutResponse, error)
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
func (s *userService) SendVerificationCode(ctx context.Context, req *dto.SendVerificationCodeRequest) (*dto.SendVerificationCodeResponse, error) {
	logger.Info(ctx, "开始处理发送验证码请求",
		logger.String("mobile", req.Mobile),
		logger.String("type", string(req.Type)))

	code := generateVerificationCode(constant.VerificationCodeLength)

	// 确定验证码类型前缀
	prefix := constant.VerificationCodePrefixLogin
	if req.Type == dto.VerificationTypeDeactivate {
		prefix = constant.VerificationCodePrefixDeactivate
	}

	// 保存验证码到Redis
	key := prefix + req.Mobile
	err := redis.Set(key, code, constant.VerificationCodeExpiration)
	if err != nil {
		logger.Error(ctx, "保存验证码到Redis失败",
			logger.String("mobile", req.Mobile),
			logger.String("type", string(req.Type)),
			logger.Err(err))
		return nil, fmt.Errorf("保存验证码失败: %w", err)
	}

	// 获取短信客户端
	client, err := sms.GetSMSClient()
	if err != nil {
		logger.Error(ctx, "创建短信客户端失败", logger.String("mobile", req.Mobile), logger.Err(err))
		return nil, fmt.Errorf("创建短信客户端失败: %w", err)
	}

	// 获取短信模板
	smsConfig := config.GetSMSConfig()
	templateCode := smsConfig.Aliyun.Templates["verification_code"]
	if templateCode == "" {
		logger.Error(ctx, "短信模板配置错误", logger.String("mobile", req.Mobile))
		return nil, fmt.Errorf("短信模板配置错误")
	}

	// 构建短信内容
	var smsContent string
	switch req.Type {
	case dto.VerificationTypeLogin:
		smsContent = fmt.Sprintf("您的登录验证码是：%s，5分钟内有效。", code)
	case dto.VerificationTypeDeactivate:
		smsContent = fmt.Sprintf("您的账号注销验证码是：%s，5分钟内有效。请谨慎操作，注销后账号将无法恢复。", code)
	default:
		smsContent = fmt.Sprintf("您的验证码是：%s，5分钟内有效。", code)
	}

	// 发送短信
	smsReq := sms.SMSRequest{
		PhoneNumbers:  req.Mobile,
		TemplateCode:  templateCode,
		TemplateParam: map[string]string{"code": code},
	}

	smsResp, err := client.SendSMS(smsReq)
	if err != nil {
		logger.Error(ctx, "发送短信失败", logger.String("mobile", req.Mobile), logger.Err(err))
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
	_ = s.smsRepo.Create(smsRecord)

	logger.Info(ctx, "验证码发送成功", logger.String("mobile", req.Mobile))
	return &dto.SendVerificationCodeResponse{Message: "验证码已发送"}, nil
}

// VerificationCodeLogin 验证码登录
func (s *userService) VerificationCodeLogin(ctx context.Context, req *dto.VerificationCodeLoginRequest) (*dto.LoginResponse, error) {
	// 记录开始处理请求的日志
	logger.Info(ctx, "开始处理验证码登录请求",
		logger.String("mobile", req.Mobile))

	// 从Redis获取验证码（登录验证码）
	key := constant.VerificationCodePrefixLogin + req.Mobile
	savedCode, err := redis.Get(key)
	if err != nil {
		logger.Error(ctx, "获取验证码失败",
			logger.String("mobile", req.Mobile),
			logger.Err(err))
		return nil, ErrInvalidCode
	}

	if savedCode != req.Code {
		logger.Warn(ctx, "验证码不匹配",
			logger.String("mobile", req.Mobile),
			logger.String("input_code", req.Code),
			logger.String("saved_code", savedCode))
		return nil, ErrInvalidCode
	}

	// 验证成功后删除验证码
	_, _ = redis.Del(key)
	logger.Debug(ctx, "验证码验证成功，已删除缓存",
		logger.String("mobile", req.Mobile))

	// 查找用户
	user, err := s.userRepo.FindByMobile(req.Mobile)
	if err != nil {
		// 如果用户不存在，则创建新用户
		logger.Info(ctx, "用户不存在，创建新用户",
			logger.String("mobile", req.Mobile))

		user = &model.User{
			Mobile:   req.Mobile,
			Username: req.Mobile,                            // 默认使用手机号作为用户名
			Nickname: "用户" + req.Mobile[len(req.Mobile)-4:], // 使用手机号后4位作为昵称
			Status:   constant.UserStatusNormal,             // 正常状态
		}

		// 保存新用户
		err = s.userRepo.Create(user)
		if err != nil {
			logger.Error(ctx, "创建用户失败",
				logger.String("mobile", req.Mobile),
				logger.Err(err))
			return nil, fmt.Errorf("创建用户失败: %w", err)
		}

		logger.Info(ctx, "新用户创建成功",
			logger.Uint("user_id", user.ID),
			logger.String("mobile", user.Mobile))
	}

	// 检查用户状态
	if user.Status != constant.UserStatusNormal {
		logger.Warn(ctx, "账号已被禁用",
			logger.Uint("user_id", user.ID),
			logger.String("mobile", user.Mobile),
			logger.Int("status", user.Status))
		return nil, errors.New("账号已被禁用")
	}

	// 生成JWT令牌
	token, err := jwt.GenerateToken(user.ID, user.Username, "")
	if err != nil {
		logger.Error(ctx, "生成令牌失败",
			logger.Uint("user_id", user.ID),
			logger.Err(err))
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

	// 登录成功
	logger.Info(ctx, "用户登录成功",
		logger.Uint("user_id", user.ID),
		logger.String("mobile", user.Mobile))

	return response, nil
}

// generateVerificationCode 生成指定长度的随机验证码
func generateVerificationCode(length int) string {
	// 使用utils包中的函数生成随机数字
	return utils.GenerateRandomDigits(length)
}

// GetUserInfo 获取用户信息
func (s *userService) GetUserInfo(ctx context.Context, id uint) (*dto.UserInfoResponse, error) {
	// 记录开始处理请求的日志
	logger.Info(ctx, "开始获取用户信息", logger.Uint("user_id", id))

	// 根据ID查找用户
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			logger.Warn(ctx, "用户不存在", logger.Uint("user_id", id))
			return nil, ErrUserNotFound
		}
		logger.Error(ctx, "查询用户失败",
			logger.Uint("user_id", id),
			logger.Err(err))
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

	// 记录成功日志
	logger.Info(ctx, "获取用户信息成功",
		logger.Uint("user_id", user.ID),
		logger.String("username", user.Username))

	return response, nil
}

// Logout 退出登录
func (s *userService) Logout(ctx context.Context, req *dto.LogoutRequest) (*dto.LogoutResponse, error) {
	// 记录开始处理请求的日志
	logger.Info(ctx, "开始处理退出登录请求")

	// 解析令牌，获取过期时间
	claims, err := jwt.ParseToken(req.Token)
	if err != nil {
		// 如果令牌已经无效，则直接返回成功
		if err == jwt.ErrTokenInvalid || err == jwt.ErrTokenExpired {
			logger.Info(ctx, "令牌已失效，无需加入黑名单")
			return &dto.LogoutResponse{Message: "退出登录成功"}, nil
		}
		logger.Error(ctx, "解析令牌失败", logger.Err(err))
		return nil, fmt.Errorf("解析令牌失败: %w", err)
	}

	// 计算令牌剩余有效期
	expTime := claims.ExpiresAt.Time
	ttl := time.Until(expTime)
	if ttl <= 0 {
		// 令牌已过期，直接返回成功
		logger.Info(ctx, "令牌已过期，无需加入黑名单")
		return &dto.LogoutResponse{Message: "退出登录成功"}, nil
	}

	// 将令牌加入黑名单，过期时间与令牌相同
	blacklistKey := TokenBlacklistPrefix + req.Token
	err = redis.Set(blacklistKey, "revoked", ttl)
	if err != nil {
		logger.Error(ctx, "将令牌加入黑名单失败",
			logger.String("token", req.Token),
			logger.Err(err))
		return nil, fmt.Errorf("退出登录失败: %w", err)
	}

	logger.Info(ctx, "用户退出登录成功",
		logger.String("username", claims.Username),
		logger.Uint("user_id", claims.UserID))

	return &dto.LogoutResponse{Message: "退出登录成功"}, nil
}

// DeactivateAccount 注销账号
func (s *userService) DeactivateAccount(ctx context.Context, req *dto.DeactivateAccountRequest) error {
	// 记录开始处理请求的日志
	logger.Info(ctx, "开始处理注销账号请求",
		logger.Uint("user_id", req.UserID),
		logger.String("mobile", req.Mobile))

	// 验证验证码（注销验证码）
	key := constant.VerificationCodePrefixDeactivate + req.Mobile
	savedCode, err := redis.Get(key)
	if err != nil {
		logger.Error(ctx, "获取注销验证码失败",
			logger.String("mobile", req.Mobile),
			logger.Err(err))
		return ErrInvalidCode
	}

	if savedCode != req.Code {
		logger.Warn(ctx, "注销验证码不匹配",
			logger.String("mobile", req.Mobile),
			logger.String("input_code", req.Code),
			logger.String("saved_code", savedCode))
		return ErrInvalidCode
	}

	// 验证成功后删除验证码
	_, _ = redis.Del(key)
	logger.Debug(ctx, "注销验证码验证成功，已删除缓存",
		logger.String("mobile", req.Mobile))

	// 查找用户
	user, err := s.userRepo.FindByID(req.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			logger.Warn(ctx, "要注销的用户不存在",
				logger.Uint("user_id", req.UserID))
			return ErrUserNotFound
		}
		logger.Error(ctx, "查询用户失败",
			logger.Uint("user_id", req.UserID),
			logger.Err(err))
		return fmt.Errorf("查询用户失败: %w", err)
	}

	// 验证手机号是否匹配
	if user.Mobile != req.Mobile {
		logger.Warn(ctx, "手机号不匹配，注销失败",
			logger.Uint("user_id", req.UserID),
			logger.String("request_mobile", req.Mobile),
			logger.String("user_mobile", user.Mobile))
		return errors.New("手机号不匹配，注销失败")
	}

	// 执行注销操作（软删除）
	err = s.userRepo.SoftDelete(req.UserID)
	if err != nil {
		logger.Error(ctx, "执行账号注销失败",
			logger.Uint("user_id", req.UserID),
			logger.Err(err))
		return ErrDeactivateFailed
	}

	logger.Info(ctx, "账号注销成功",
		logger.Uint("user_id", req.UserID),
		logger.String("mobile", user.Mobile))

	return nil
}
