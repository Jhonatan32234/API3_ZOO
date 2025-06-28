package routes

import (
	"api3/src/controllers"
	"api3/src/utils"

	"github.com/gorilla/mux"
)

func SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/login", controllers.Login).Methods("POST")

	r.HandleFunc("/register", /*utils.RequireRole("dev")*/(controllers.Register)).Methods("POST")
	r.HandleFunc("/update/{id}", utils.RequireRole("admin")(controllers.UpdateUser)).Methods("PUT")
	r.HandleFunc("/delete/{id}", utils.RequireRole("admin")(controllers.DeleteUser)).Methods("DELETE")
	r.HandleFunc("/users", (controllers.GetAllUsers)).Methods("GET")


	return r
}
