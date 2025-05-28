package repository

import (
	"app/internal/constant"
	"app/internal/model"
	"fmt"

	"gorm.io/gorm"
)

// PostRepository 动态仓库接口
type PostRepository interface {
	// 查询方法
	GetPost(id uint) (*model.Post, error)
	GetUserPosts(userID uint, page, size int, viewerID ...uint) ([]model.Post, int64, error)
	GetFollowingPosts(userID uint, page, size int) ([]model.Post, int64, error)

	// 修改方法
	CreatePost(post *model.Post) error
	UpdatePost(post *model.Post) error
	IncrementPostLikes(postID uint) error
	IncrementPostComments(postID uint) error
	// 事务方法
	IncrementPostCommentsWithTx(tx *gorm.DB, postID uint) error
}

// postRepository 动态仓库实现
type postRepository struct {
	db *gorm.DB
}

// NewPostRepository 创建动态仓库实例
func NewPostRepository(db *gorm.DB) PostRepository {
	return &postRepository{db: db}
}

// GetPost 获取动态
func (r *postRepository) GetPost(id uint) (*model.Post, error) {
	var post model.Post
	err := r.db.First(&post, id).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

// GetUserPosts 获取用户动态列表
func (r *postRepository) GetUserPosts(userID uint, page, size int, viewerID ...uint) ([]model.Post, int64, error) {
	var posts []model.Post
	var count int64

	offset := (page - 1) * size

	// 基础查询：获取指定用户的动态
	query := r.db.Model(&model.Post{}).Where("user_id = ?", userID)

	// 如果提供了查看者ID且不是自己查看自己的动态，需要根据可见性过滤
	if len(viewerID) > 0 && viewerID[0] != userID {
		// 检查是否为好友关系（双记录模式）
		var friendCount int64
		r.db.Model(&model.UserFriend{}).
			Where("user_id = ? AND target_id = ? AND status = ? AND direction IN (0, 1)", viewerID[0], userID, int(constant.FriendStatusConfirmed)).
			Count(&friendCount)

		if friendCount > 0 {
			// 是好友关系，可以看到公开和好友可见的动态
			query = query.Where("visibility IN (?, ?)", int(constant.VisibilityPublic), int(constant.VisibilityFriends))
		} else {
			// 不是好友关系，只能看到公开动态
			query = query.Where("visibility = ?", int(constant.VisibilityPublic))
		}
	}

	// 计算总数
	err := query.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err = query.Order("created_at DESC").Offset(offset).Limit(size).Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	return posts, count, nil
}

// GetFollowingPosts 获取关注用户的动态列表
func (r *postRepository) GetFollowingPosts(userID uint, page, size int) ([]model.Post, int64, error) {
	var posts []model.Post
	var count int64

	offset := (page - 1) * size

	// 构建复杂查询
	// 1. 获取所有关注用户的公开动态
	publicPostsQuery := r.db.Table("posts").
		Select("posts.*").
		Joins("JOIN user_follower ON posts.user_id = user_follower.target_id").
		Where("user_follower.user_id = ?", userID).
		Where("posts.visibility = ?", int(constant.VisibilityPublic))

	// 2. 获取好友的仅好友可见动态
	friendPostsQuery := r.db.Table("posts").
		Select("posts.*").
		Joins("JOIN user_friend ON posts.user_id = user_friend.target_id").
		Where("user_friend.user_id = ?", userID).
		Where("user_friend.status = ? AND user_friend.direction IN (0, 1)", int(constant.FriendStatusConfirmed)). // 已确认的好友关系（双记录模式）
		Where("posts.visibility = ?", int(constant.VisibilityFriends))

	// 使用UNION合并查询结果
	// 获取公开动态的SQL
	publicSQL := publicPostsQuery.Session(&gorm.Session{DryRun: true}).Find(&[]model.Post{}).Statement.SQL.String()
	publicVars := publicPostsQuery.Session(&gorm.Session{DryRun: true}).Find(&[]model.Post{}).Statement.Vars

	// 获取好友动态的SQL
	friendSQL := friendPostsQuery.Session(&gorm.Session{DryRun: true}).Find(&[]model.Post{}).Statement.SQL.String()
	friendVars := friendPostsQuery.Session(&gorm.Session{DryRun: true}).Find(&[]model.Post{}).Statement.Vars

	// 合并变量
	allVars := append(publicVars, friendVars...)

	// 构建UNION查询
	unionSQL := fmt.Sprintf("(%s) UNION (%s)", publicSQL, friendSQL)

	// 计算总数
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS count_table", unionSQL)
	err := r.db.Raw(countSQL, allVars...).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	resultSQL := fmt.Sprintf("SELECT * FROM (%s) AS combined_posts ORDER BY created_at DESC LIMIT %d OFFSET %d", unionSQL, size, offset)
	err = r.db.Raw(resultSQL, allVars...).Scan(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	return posts, count, nil
}

// CreatePost 创建动态
func (r *postRepository) CreatePost(post *model.Post) error {
	return r.db.Create(post).Error
}

// IncrementPostLikes 增加动态点赞数
func (r *postRepository) IncrementPostLikes(postID uint) error {
	return r.db.Model(&model.Post{}).Where("id = ?", postID).Update("likes", gorm.Expr("likes + ?", 1)).Error
}

// UpdatePost 更新动态信息
func (r *postRepository) UpdatePost(post *model.Post) error {
	return r.db.Save(post).Error
}

// IncrementPostComments 增加动态评论数
func (r *postRepository) IncrementPostComments(postID uint) error {
	return r.db.Model(&model.Post{}).Where("id = ?", postID).Update("comments", gorm.Expr("comments + ?", 1)).Error
}

// IncrementPostCommentsWithTx 在事务中增加动态评论数
func (r *postRepository) IncrementPostCommentsWithTx(tx *gorm.DB, postID uint) error {
	return tx.Model(&model.Post{}).Where("id = ?", postID).Update("comments", gorm.Expr("comments + ?", 1)).Error
}
