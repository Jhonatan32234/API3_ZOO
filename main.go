package main

import (
	"api3/db"
	"api3/src/routes"
	"log"
	"net/http"
)

func main() {
	db.ConnectDB()
	r := routes.SetupRoutes()

	log.Println("✅ Servidor corriendo en :8082")
	log.Fatal(http.ListenAndServe(":8082", r))
}
