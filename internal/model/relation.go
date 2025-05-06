package model

import (
	"time"

	"gorm.io/gorm"
)

// Relation 关系模型
// 存储用户之间的关系，如关注、好友等
type Relation struct {
	ID        uint           `gorm:"primaryKey;comment:关系ID，主键" json:"id"`
	UserID    uint           `gorm:"index:idx_user_relation,priority:1;index:idx_user_target,priority:1,uniqueIndex;comment:用户ID，关系发起者" json:"user_id"`
	TargetID  uint           `gorm:"index:idx_user_relation,priority:2;index:idx_user_target,priority:2,uniqueIndex;comment:目标用户ID，关系接收者" json:"target_id"`
	Type      int            `gorm:"type:smallint;default:1;comment:关系类型：1-关注，2-好友" json:"type"`
	Status    int            `gorm:"type:smallint;default:1;comment:关系状态：0-待确认，1-已确认" json:"status"`
	CreatedAt time.Time      `gorm:"type:datetime;comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"type:datetime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"type:datetime;comment:删除时间，软删除" json:"-"`
}
