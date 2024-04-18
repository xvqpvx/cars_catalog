package service

import (
	"car_catalog/internal/dto"
	"context"
)

type CarService interface {
	GetFilteredCars(ctx context.Context, filters dto.Filters, cursors dto.Cursors) ([]dto.GetFilteredCarsDto, dto.Cursors, error)
	DeleteCar(ctx context.Context, carId string) error
	UpdateCar(ctx context.Context, carId string, car dto.UpdateCarDto) error
	AddCars(ctx context.Context, cars []dto.AddCarsDto) error
}
