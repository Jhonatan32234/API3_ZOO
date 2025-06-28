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

// Register godoc
// @Summary Registrar nuevo usuario
// @Description Crea un nuevo usuario en la base de datos
// @Tags users
// @Accept json
// @Produce plain
// @Param user body models.User true "Datos del nuevo usuario"
// @Success 201 {string} string "Usuario creado"
// @Failure 400 {string} string "Error al registrar usuario"
// @Router /register [post]
func Register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // máximo 10MB
	if err != nil {
		http.Error(w, "No se pudo parsear el formulario", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	role := r.FormValue("role")
	if role == "" {
		role = "user"
	}

	// Manejo de imagen
	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error al recibir la imagen", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Guarda la imagen en disco
	imagePath := "uploads/" + handler.Filename
	dst, err := utils.SaveFile(file, imagePath)
	if err != nil {
		http.Error(w, "No se pudo guardar la imagen", http.StatusInternalServerError)
		return
	}
	if dst == "" {
		http.Error(w, "No se pudo guardar la imagen", http.StatusInternalServerError)
		return
	}

	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := models.User{
		Username: username,
		Password: string(hashedPwd),
		Role:     role,
		Image:    imagePath,
	}

	result := db.DB.Create(&user)
	if result.Error != nil {
		http.Error(w, "Error al registrar usuario", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Usuario creado"))
}


// Login godoc
// @Summary Iniciar sesión
// @Description Autentica un usuario y devuelve un token JWT
// @Tags auth
// @Accept json
// @Produce json
// @Param user body models.User true "Credenciales de usuario"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /login [post]
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


// GetAllUsers godoc
// @Summary Obtener todos los usuarios
// @Description Retorna todos los usuarios registrados (requiere rol admin)
// @Tags users
// @Produce json
// @Success 200 {array} models.User
// @Failure 500 {object} map[string]string
// @Router /users [get]
func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User

	if err := db.DB.Find(&users).Error; err != nil {
		http.Error(w, "Error al obtener usuarios", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}



// UpdateUser godoc
// @Summary Actualizar usuario
// @Description Actualiza los datos de un usuario existente (requiere rol dev)
// @Tags users
// @Accept json
// @Produce plain
// @Param id path int true "ID del usuario"
// @Param user body models.User true "Datos actualizados"
// @Success 200 {string} string "Usuario actualizado"
// @Failure 404 {string} string "Usuario no encontrado"
// @Router /update/{id} [put]
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

// DeleteUser godoc
// @Summary Eliminar usuario
// @Description Elimina un usuario de la base de datos (requiere rol dev)
// @Tags users
// @Produce plain
// @Param id path int true "ID del usuario"
// @Success 200 {string} string "Usuario eliminado"
// @Failure 500 {string} string "Error al eliminar usuario"
// @Router /delete/{id} [delete]
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	idParam := mux.Vars(r)["id"]
	id, _ := strconv.Atoi(idParam)

	if err := db.DB.Delete(&models.User{}, id).Error; err != nil {
		http.Error(w, "Error al eliminar usuario", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Usuario eliminado"))
}
