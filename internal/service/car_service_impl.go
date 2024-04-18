package service

import (
	"car_catalog/internal/dto"
	"car_catalog/internal/model"
	"car_catalog/internal/repository"
	"context"
	"encoding/base64"
	"log"
	"strconv"
)

type CarServiceImpl struct {
	CarRepo repository.CarRepository
}

func NewCarService(carRepo repository.CarRepository) CarService {
	return &CarServiceImpl{CarRepo: carRepo}
}

func (c *CarServiceImpl) AddCars(ctx context.Context, cars []dto.AddCarsDto) error {
	var carsToAdd []model.Car

	for _, car := range cars {
		log.Printf("[DEBUG] Service - AddCars - Adding car: %+v", car)
		carToAdd := model.Car{
			Mark:            car.Mark,
			Model:           car.Model,
			Year:            car.Year,
			RegNum:          car.RegNum,
			OwnerName:       car.Owner.Name,
			OwnerSurname:    car.Owner.Surname,
			OwnerPatronymic: car.Owner.Patronymic,
		}
		carsToAdd = append(carsToAdd, carToAdd)
	}

	if err := c.CarRepo.AddCars(ctx, carsToAdd); err != nil {
		log.Printf("[ERROR] Service - AddCars - Error adding cars: %v", err)
		return err
	}

	log.Println("[INFO] Service - AddCars - Cars added successfully")
	return nil
}

func (c *CarServiceImpl) DeleteCar(ctx context.Context, carId string) error {
	carID, err := strconv.Atoi(carId)
	if err != nil {
		log.Printf("[ERROR] Service - UpdateCar - Unable to parse car id, error: %v", err)
		return err
	}
	log.Printf("[DEBUG] Service - DeleteCar - Car ID to delete: %d", carID)

	carToDelete, err := c.CarRepo.GetCarById(ctx, carID)
	if err != nil {
		log.Printf("[ERROR] Service - DeleteCar - Error getting car with id %s: %v", carId, err)
		return err
	}

	if err := c.CarRepo.DeleteCar(ctx, carToDelete.CarId); err != nil {
		log.Printf("[ERROR] Service - DeleteCar - Error deleting car: %v", err)
		return err
	}

	log.Println("[INFO] Service - DeleteCar - Car deleted successfully")
	return nil
}

func DecodeCursor(cursors *dto.Cursors) error {
	decoded, err := base64.StdEncoding.DecodeString(cursors.Next)
	if err != nil {
		log.Printf("Unable to decode next cursor err: %v", err)
		return err
	}
	(*cursors).Next = string(decoded)
	decoded, err = base64.StdEncoding.DecodeString(cursors.Prev)
	if err != nil {
		log.Printf("Unable to decode prev cursor err: %v", err)
		return err
	}
	(*cursors).Prev = string(decoded)

	return nil
}

func EncodeCursor(cursors *dto.Cursors) {
	(*cursors).Next = base64.StdEncoding.EncodeToString([]byte(cursors.Next))
	(*cursors).Prev = base64.StdEncoding.EncodeToString([]byte(cursors.Prev))
}

func (c *CarServiceImpl) GetFilteredCars(ctx context.Context, filters dto.Filters, cursors dto.Cursors) ([]dto.GetFilteredCarsDto, dto.Cursors, error) {
	limit, err := strconv.Atoi(filters.Limit)
	if err != nil {
		log.Printf("[ERROR] Service - GetFilteredCars - Unable to parse limit error: %v", err)
		return []dto.GetFilteredCarsDto{}, dto.Cursors{}, err
	}

	if err := DecodeCursor(&cursors); err != nil {
		log.Printf("[ERROR] Service - GetFilteredCars - error: %v", err)
		return []dto.GetFilteredCarsDto{}, dto.Cursors{}, err
	}

	carsToFilter, cursors, err := c.CarRepo.GetCars(ctx, limit, filters.Mark, filters.Model, filters.Year, cursors)
	if err != nil {
		log.Printf("[ERROR] Service - GetFilteredCars - Error getting all cars: %v", err)
		return []dto.GetFilteredCarsDto{}, dto.Cursors{}, err
	}

	var filteredCars []dto.GetFilteredCarsDto
	for _, car := range carsToFilter {
		log.Printf("[DEBUG] Service - GetFilteredCars - Adding car: %+v", car)
		filteredCar := dto.GetFilteredCarsDto{
			CarId: car.CarId,
			Mark:  car.Mark,
			Model: car.Model,
			Year:  strconv.Itoa(car.Year),
		}
		filteredCars = append(filteredCars, filteredCar)
	}

	EncodeCursor(&cursors)

	log.Println("[INFO] Service - GetFilteredCars - Cars filtered successfully")
	return filteredCars, cursors, nil
}

func (c *CarServiceImpl) UpdateCar(ctx context.Context, carId string, car dto.UpdateCarDto) error {
	carID, err := strconv.Atoi(carId)
	if err != nil {
		log.Printf("[ERROR] Service - UpdateCar - Unable to parse car id, error: %v", err)
		return err
	}
	log.Printf("[DEBUG] Service - UpdateCar - Car ID to update: %d", carID)

	carToUpdate, err := c.CarRepo.GetCarById(ctx, carID)
	if err != nil {
		log.Printf("[ERROR] Service - UpdateCar - Error getting car with id %d: %v", carID, err)
		return err
	}

	if car.Mark != "" {
		log.Printf("[DEBUG] Service - UpdateCar - Updating car mark to: %s", car.Mark)
		carToUpdate.Mark = car.Mark
	}
	if car.Model != "" {
		log.Printf("[DEBUG] Service - UpdateCar - Updating car model to: %s", car.Model)
		carToUpdate.Model = car.Model
	}
	if car.Year != "" {
		log.Printf("[DEBUG] Service - UpdateCar - Updating car year to: %s", car.Year)
		year, err := strconv.Atoi(car.Year)
		if err != nil {
			log.Printf("[ERROR] Service - UpdateCar - Unable to parse car year, error: %v", err)
		} else {
			carToUpdate.Year = year
		}
	}
	if car.RegNum != "" {
		log.Printf("[DEBUG] Service - UpdateCar - Updating car regNum to: %s", car.RegNum)
		carToUpdate.RegNum = car.RegNum
	}

	if err := c.CarRepo.UpdateCar(ctx, carToUpdate); err != nil {
		log.Printf("[ERROR] Service - Update car - Error updating car fields: %v", err)
		return err
	}

	log.Println("[INFO] Service - UpdateCar - Car updated successfully")
	return nil
}
