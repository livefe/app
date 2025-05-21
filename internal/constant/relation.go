package constant

// SocialRelationType 社交关系类型
type SocialRelationType int

const (
	// 关注关系类型
	SocialRelationTypeFollower SocialRelationType = 1
	// 好友关系类型
	SocialRelationTypeFriend SocialRelationType = 2
)

// FriendStatus 好友关系状态
type FriendStatus int

const (
	// 好友请求待确认状态
	FriendStatusPending FriendStatus = 0
	// 好友关系已确认状态
	FriendStatusConfirmed FriendStatus = 1
)

// Visibility 内容可见性类型
type Visibility int

const (
	// 公开可见
	VisibilityPublic Visibility = 1
	// 仅好友可见
	VisibilityFriends Visibility = 2
	// 私密可见
	VisibilityPrivate Visibility = 3
)
