package main

import (
	"car_catalog/internal/pkg/app"
	"log"
)

// @title Cars Catalog API
// @version 1.0
// @description API Server for Cars Catalog Application

// @host localhost:8080
// @BasePath /

func main() {
	a, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	err = a.Run()
	if err != nil {
		log.Fatal(err)
	}
}
