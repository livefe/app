package repository

import (
	"app/internal/model"

	"gorm.io/gorm"
)

// SMSRepository SMS记录仓库接口
type SMSRepository interface {
	// Create 创建SMS记录
	Create(record *model.SMSRecord) error
	// FindByPhoneNumber 根据手机号查找SMS记录
	FindByPhoneNumber(phoneNumber string, limit int) ([]*model.SMSRecord, error)
	// FindByID 根据ID查找SMS记录
	FindByID(id uint) (*model.SMSRecord, error)
}

// smsRepository SMS记录仓库实现
type smsRepository struct {
	db *gorm.DB
}

// NewSMSRepository 创建SMS记录仓库实例
func NewSMSRepository(db *gorm.DB) SMSRepository {
	return &smsRepository{
		db: db,
	}
}

// Create 创建SMS记录
func (r *smsRepository) Create(record *model.SMSRecord) error {
	result := r.db.Create(record)
	return result.Error
}

// FindByPhoneNumber 根据手机号查找SMS记录
func (r *smsRepository) FindByPhoneNumber(phoneNumber string, limit int) ([]*model.SMSRecord, error) {
	var records []*model.SMSRecord
	result := r.db.Where("phone_number = ?", phoneNumber).Order("created_at DESC").Limit(limit).Find(&records)
	if result.Error != nil {
		return nil, result.Error
	}
	return records, nil
}

// FindByID 根据ID查找SMS记录
func (r *smsRepository) FindByID(id uint) (*model.SMSRecord, error) {
	var record model.SMSRecord
	result := r.db.First(&record, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, ErrRecordNotFound
		}
		return nil, result.Error
	}
	return &record, nil
}
