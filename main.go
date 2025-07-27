package main

import (
	"api3/db"
	_ "api3/docs"
	"api3/src/routes"
	"api3/src/utils"
	"log"
	"github.com/joho/godotenv"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	err := godotenv.Load(".env")
    if err != nil {
        log.Println("Advertencia: no se pudo cargar el archivo .env:", err)
    }
	db.ConnectDB()
	r := routes.SetupRoutes()
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	handlerWithCORS := utils.CORS(r)


	log.Println("✅ Servidor corriendo en :8080")
	log.Fatal(http.ListenAndServe(":8080", handlerWithCORS))
}
