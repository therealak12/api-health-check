package repository

import (
	"errors"
	"gorm.io/gorm"
)

type Healthcheck struct {
	ID              int    `json:"id"`
	IntervalSeconds int    `json:"intervalSeconds"`
	Url             string `json:"url"`
	HttpMethod      string `json:"httpMethod"`
	HeadersJson     string `json:"headers"`
	Body            string `json:"body"`
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

	if errors.Is(query.Error, gorm.ErrRecordNotFound) || query.RowsAffected == 0 {
		return healthcheck, ErrRecordNotFound
	}
	if query.Error != nil {
		return healthcheck, query.Error
	}

	return healthcheck, nil
}

func (c SQLHealthcheckRepo) Delete(id int) error {
	query := c.DB.Where("id = ?", id).Delete(&Healthcheck{})

	if errors.Is(query.Error, gorm.ErrRecordNotFound) || query.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	if query.Error != nil {
		return query.Error
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
