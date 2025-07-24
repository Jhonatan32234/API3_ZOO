package main

import (
	"api3/db"
	_ "api3/docs"
	"api3/src/routes"
	"api3/src/utils"
	"log"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	db.ConnectDB()
	r := routes.SetupRoutes()
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	handlerWithCORS := utils.CORS(r)


	log.Println("✅ Servidor corriendo en :8080")
	log.Fatal(http.ListenAndServe(":8080", handlerWithCORS))
}
