package controllers

import (
	"api3/db"
	"api3/src/models"
	"api3/src/utils"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

func Register(w http.ResponseWriter, r *http.Request) {
	var user models.User
	json.NewDecoder(r.Body).Decode(&user)

	if user.Role == "" {
		user.Role = "user"
	}

	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(hashedPwd)

	result := db.DB.Create(&user)
	if result.Error != nil {
		http.Error(w, "Error al registrar usuario", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Usuario creado"))
}

func Login(w http.ResponseWriter, r *http.Request) {
	var input models.User
	json.NewDecoder(r.Body).Decode(&input)

	var dbUser models.User
	result := db.DB.Where("username = ?", input.Username).First(&dbUser)
	if result.Error != nil {
		http.Error(w, "Usuario no encontrado", http.StatusUnauthorized)
		return
	}

	err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(input.Password))
	if err != nil {
		http.Error(w, "Contraseña incorrecta", http.StatusUnauthorized)
		return
	}

	token, err := utils.GenerateToken(dbUser.ID, dbUser.Role)
	if err != nil {
		http.Error(w, "No se pudo generar token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	idParam := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idParam)

	var user models.User
	if err := db.DB.First(&user, id).Error; err != nil {
		http.Error(w, "Usuario no encontrado", http.StatusNotFound)
		return
	}

	var updateData models.User
	json.NewDecoder(r.Body).Decode(&updateData)

	if updateData.Password != "" {
		hashed, _ := bcrypt.GenerateFromPassword([]byte(updateData.Password), bcrypt.DefaultCost)
		updateData.Password = string(hashed)
	}

	db.DB.Model(&user).Updates(updateData)
	w.Write([]byte("Usuario actualizado"))
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	idParam := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idParam)

	if err := db.DB.Delete(&models.User{}, id).Error; err != nil {
		http.Error(w, "Error al eliminar usuario", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Usuario eliminado"))
}
