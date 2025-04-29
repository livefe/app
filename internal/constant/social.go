package constant

// 社交关系类型常量
const (
	// RelationTypeFollow 关注关系
	RelationTypeFollow = 1
	// RelationTypeFriend 好友关系
	RelationTypeFriend = 2
)

// 社交关系状态常量
const (
	// RelationStatusPending 待确认状态
	RelationStatusPending = 0
	// RelationStatusConfirmed 已确认状态
	RelationStatusConfirmed = 1
)

// 位置分享可见性常量
const (
	// VisibilityPublic 公开可见
	VisibilityPublic = 1
	// VisibilityFriends 仅好友可见
	VisibilityFriends = 2
	// VisibilityPrivate 私密可见
	VisibilityPrivate = 3
)
