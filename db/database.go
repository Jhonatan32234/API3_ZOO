package db

import (
	"api3/src/models"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	//dsn := "root:root@tcp(localhost:3306)/userdb?parseTime=true"
	//dsn := "root:root@tcp(mysql-container:3306)/userdb?parseTime=true"
	dsn := "root:root@tcp(34.229.32.55:3306)/userdb?parseTime=true"

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Error al conectar con la BD:", err)
	}

	err = DB.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal("❌ Error al migrar modelo User:", err)
	}

	fmt.Println("✅ Conectado a MySQL y tabla 'users' lista")
}
