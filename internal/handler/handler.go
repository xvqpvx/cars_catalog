package handler

import (
	"car_catalog/internal/dto"
	"car_catalog/internal/service"
	"encoding/json"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type CarHandler struct {
	CarService     service.CarService
	ExternalAPIURL string
}

func NewCarHandler(carService service.CarService, externalAPIURL string) *CarHandler {
	return &CarHandler{
		CarService:     carService,
		ExternalAPIURL: externalAPIURL,
	}
}

// @Summary Get cars list
// @Description Get cars list by filters with pagination
// @Tags cars
// @Produce  json
// @Param mark query string false "Car mark"
// @Param model query string false "Car model"
// @Param year query string false "Car year"
// @Param limit query string false "Results limit" default(10)
// @Param next query string false "Next cursor for pagination"
// @Param prev query string false "Previous cursor for pagination"
// @Success 200 {object} map[string]interface{} "OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 405 {string} string "Method Not Allowed"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/getCars [get]
func (c *CarHandler) GetFilteredCars(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	log.Println("[INFO] Handler - GetFilteredCars - Received GET request")

	filters := dto.Filters{
		Mark:  r.URL.Query().Get("mark"),
		Model: r.URL.Query().Get("model"),
		Year:  r.URL.Query().Get("year"),
		Limit: r.URL.Query().Get("limit"),
	}
	log.Printf("[DEBUG] Handler - GetFilteredCars - Filters: %+v", filters)

	cursors := dto.Cursors{
		Next: r.URL.Query().Get("next"),
		Prev: r.URL.Query().Get("prev"),
	}
	log.Printf("[DEBUG] Handler - GetFilteredCars - Cursors: %+v", cursors)

	result, cursors, err := c.CarService.GetFilteredCars(r.Context(), filters, cursors)
	if err != nil {
		log.Printf("[ERROR] Handler - GetFilteredCars - Unable to get filtered cars error: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"cursors": cursors,
		"cars":    result,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Printf("[ERROR] Handler - GetFilteredCars - Unable to encode JSON: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

// @Summary Add cars
// @Description Add cars to the database from external API based on registration numbers
// @Tags cars
// @Accept  json
// @Produce  json
// @Param regNums body dto.RegNumsRequest true "Registration numbers array"
// @Success 200 {string} string "Request processed successfully"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/addCars [post]
func (c *CarHandler) AddCars(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	log.Println("[INFO] Handler - AddCars - Received POST request")

	var requestBody dto.RegNumsRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		log.Printf("[ERROR] Handler - AddCars - Unable to decode JSON: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	cars := []dto.AddCarsDto{}
	for _, regNum := range requestBody.RegNums {
		var carResp dto.AddCarsDto

		resp, err := http.Get(c.ExternalAPIURL + "?regNum=" + regNum)
		if err != nil {
			log.Printf("[ERROR] Handler - AddCars - Failed to get car information from external API: %v", err)
			http.Error(w, "Failed to get car information from external API", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusBadRequest {
			log.Printf("[INFO] Handler - AddCars - Got StatusBadRequest from external API")
			http.Error(w, "StatusBadRequest from external API", http.StatusInternalServerError)
			return
		} else if resp.StatusCode == http.StatusInternalServerError {
			log.Printf("[INFO] Handler - AddCars - Got StatusInternalServerError from external API")
			http.Error(w, "StatusInternalServerError from external API", http.StatusInternalServerError)
			return
		} else if resp.StatusCode == http.StatusOK {
			if err := json.NewDecoder(resp.Body).Decode(&carResp); err != nil {
				log.Printf("[ERROR] Handler - AddCars - Unable to decode response from external API: %v", err)
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
			cars = append(cars, carResp)
		}
	}

	if err := c.CarService.AddCars(r.Context(), cars); err != nil {
		log.Printf("[ERROR] Handler - AddCars - Unable to add car: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Request processed successfully"))
}

// @Summary Update a car
// @Description Update a car by its ID
// @Tags cars
// @Accept json
// @Produce json
// @Param id path string true "Car ID"
// @Param updateDto body dto.UpdateCarDto true "Car update information"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/updateCar/{id} [patch]
func (c *CarHandler) UpdateCar(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	log.Printf("[INFO] Handler - UpdateCar - Received PATCH request for car ID: %s", p.ByName("id"))

	carId := p.ByName("id")
	var updateDto dto.UpdateCarDto
	if err := json.NewDecoder(r.Body).Decode(&updateDto); err != nil {
		log.Printf("[ERROR] Handler - UpdateCar - Unable to decode JSON: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if err := c.CarService.UpdateCar(r.Context(), carId, updateDto); err != nil {
		log.Printf("[ERROR] Handler - UpdateCar - Unable to update car error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Request processed successfully"))
}

// @Summary Delete a car
// @Description Delete a car by its ID
// @Tags cars
// @Param id path string true "Car ID"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/delete/{id} [delete]
func (c *CarHandler) DeleteCar(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	log.Printf("[INFO] Handler - DeleteCar - Received DELETE request for car ID: %s", p.ByName("id"))

	carId := p.ByName("id")
	if err := c.CarService.DeleteCar(r.Context(), carId); err != nil {
		log.Printf("[ERROR] Handler - DeleteCar - Unable to delete car: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Request processed successfully"))
}
