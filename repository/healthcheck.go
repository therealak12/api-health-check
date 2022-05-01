package repository

import (
	"fmt"

	"gorm.io/gorm"
)

type Healthcheck struct {
	ID              int      `json:"id"`
	IntervalSeconds int      `json:"IntervalSeconds"`
	Url             string   `json:"url"`
	HttpMethod      string   `json:"httpMethod"`
	Headers         []string `json:"headers"`
	Body            string   `json:"body"`
}

type HealthcheckRepo interface {
	Delete(id int) error
	Save(healthcheck *Healthcheck) error
	FindOne(id int) (Healthcheck, error)
	FindAll() ([]Healthcheck, error)
}

var _ HealthcheckRepo = SQLHealthcheckRepo{}

type SQLHealthcheckRepo struct {
	DB *gorm.DB
}

func (c SQLHealthcheckRepo) FindOne(id int) (Healthcheck, error) {
	healthcheck := Healthcheck{}
	query := c.DB.Where("id = ?", id).Find(&healthcheck)
	if query.Error != nil {
		return healthcheck, query.Error
	}

	if query.RowsAffected == 0 {
		return healthcheck, fmt.Errorf("sql healthcheck find one: %w", ErrRecordNotFound)
	}

	return healthcheck, nil
}

func (c SQLHealthcheckRepo) Delete(id int) error {
	query := c.DB.Where("id = ?", id).Delete(&Healthcheck{})
	if query.Error != nil {
		return query.Error
	}

	if query.RowsAffected == 0 {
		return fmt.Errorf("sql healthcheck delete: %w", ErrRecordNotFound)
	}

	return nil
}

func (c SQLHealthcheckRepo) Save(healthcheck *Healthcheck) error {
	return c.DB.Save(healthcheck).Error
}

func (c SQLHealthcheckRepo) FindAll() ([]Healthcheck, error) {
	var result []Healthcheck
	err := c.DB.Find(&result).Error

	return result, err
}
