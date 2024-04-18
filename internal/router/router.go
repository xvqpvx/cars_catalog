package router

import (
	_ "car_catalog/docs"
	"car_catalog/internal/handler"
	"net/http"

	"github.com/julienschmidt/httprouter"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(carHandler *handler.CarHandler) *httprouter.Router {
	router := httprouter.New()

	router.GET("/api/getCars/", carHandler.GetFilteredCars)
	router.POST("/api/addCars", carHandler.AddCars)
	router.PATCH("/api/updateCar/:id", carHandler.UpdateCar)
	router.DELETE("/api/delete/:id", carHandler.DeleteCar)

	router.GET("/swagger/*any", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		httpSwagger.WrapHandler(w, r)
	})
	return router
}
