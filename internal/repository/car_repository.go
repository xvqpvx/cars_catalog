package repository

import (
	"car_catalog/internal/model"
	"car_catalog/internal/dto"
	"context"
)

type CarRepository interface {
	AddCars(ctx context.Context, cars []model.Car) error
	GetCarById(ctx context.Context, carId int) (model.Car, error)
	GetCars(ctx context.Context, limit int, mark, carModel, year string, cursors dto.Cursors) ([]model.Car, dto.Cursors, error)
	UpdateCar(ctx context.Context, car model.Car) error
	DeleteCar(ctx context.Context, carId int) error
}
