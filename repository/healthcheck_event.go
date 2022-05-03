package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type HealthcheckEvent struct {
	ID            int       `json:"id"`
	HealthcheckID int       `json:"-"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
}

type HealthcheckEventRepo interface {
	Create(healthcheckEvent *HealthcheckEvent) error
	FindLast() (HealthcheckEvent, error)
}

var _ HealthcheckEventRepo = SQLHealthcheckEventRepo{}

type SQLHealthcheckEventRepo struct {
	DB *gorm.DB
}

func (c SQLHealthcheckEventRepo) Create(event *HealthcheckEvent) error {
	return c.DB.Save(event).Error
}

func (c SQLHealthcheckEventRepo) FindLast() (HealthcheckEvent, error) {
	event := HealthcheckEvent{}
	query := c.DB.Last(&event)
	if query.Error != nil {
		if errors.Is(query.Error, gorm.ErrRecordNotFound) || query.RowsAffected == 0 {
			return event, ErrRecordNotFound
		}
		return event, query.Error
	}

	return event, nil
}
