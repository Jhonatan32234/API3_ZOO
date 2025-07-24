package controllers

import (
	"api3/db"
	"api3/src/models"
	"api3/src/utils"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

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
	var username, password, role, zona string
	var imageBytes []byte

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, "Error al parsear formulario: "+err.Error(), http.StatusBadRequest)
			return
		}

		username = r.FormValue("username")
		password = r.FormValue("password")
		role = r.FormValue("role")
		zona = r.FormValue("zona")
		if role == "" {
			role = "user"
		}

		file, _, err := r.FormFile("image")
		if err == nil {
			defer file.Close()
			var buf bytes.Buffer
			io.Copy(&buf, file)
			imageBytes = buf.Bytes()
		}

	} else {
		var input struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Role     string `json:"role"`
			Zona     string `json:"zona"`
			Image    string `json:"image"` // base64
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Error de JSON", http.StatusBadRequest)
			return
		}
		username = input.Username
		password = input.Password
		role = input.Role
		zona = input.Zona

		if input.Image != "" {
			imageBytes, _ = base64.StdEncoding.DecodeString(input.Image)
		}
	}

	if username == "" || password == "" || zona == "" {
		http.Error(w, "Faltan campos obligatorios", http.StatusBadRequest)
		return
	}

	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := models.User{
		Username: username,
		Password: string(hashedPwd),
		Role:     role,
		Zona:     zona,
		Image:    imageBytes,
	}

	if err := db.DB.Create(&user).Error; err != nil {
		http.Error(w, "Error al guardar usuario", http.StatusBadRequest)
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
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	if input.Username == "" || input.Password == "" {
		http.Error(w, "Username y password son obligatorios", http.StatusBadRequest)
		return
	}

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

	token, err := utils.GenerateToken(uint(dbUser.ID), dbUser.Role, dbUser.Zona)
	if err != nil {
		http.Error(w, "No se pudo generar token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token":    token,
		"username": dbUser.Username,
		"role":     dbUser.Role,
		"zona":     dbUser.Zona,
	})
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

	// Formatear imágenes
	for i := range users {
		users[i].FormatImage()
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
	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var user models.User
	if err := db.DB.First(&user, id).Error; err != nil {
		http.Error(w, "Usuario no encontrado", http.StatusNotFound)
		return
	}

	contentType := r.Header.Get("Content-Type")

	var username, role, password, zona string
	var imageBase64 string
	imageUpdated := false

	if strings.HasPrefix(contentType, "multipart/form-data") {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, "No se pudo parsear el formulario: "+err.Error(), http.StatusBadRequest)
			return
		}

		username = r.FormValue("username")
		role = r.FormValue("role")
		password = r.FormValue("password")
		zona = r.FormValue("zona")

		file, _, err := r.FormFile("image")
		if err == nil {
			defer file.Close()

			var buf bytes.Buffer
			if _, err := io.Copy(&buf, file); err != nil {
				http.Error(w, "No se pudo leer la imagen: "+err.Error(), http.StatusInternalServerError)
				return
			}
			imageBase64 = base64.StdEncoding.EncodeToString(buf.Bytes())
			imageUpdated = true
		} else {
			if err != http.ErrMissingFile {
				http.Error(w, "Error al procesar imagen: "+err.Error(), http.StatusBadRequest)
				return
			}
		}
	} else {
		var input struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Role     string `json:"role"`
			Zona     string `json:"zona"`
			Image    string `json:"image"` // Para actualizar imagen desde JSON
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Error en el formato JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		username = input.Username
		password = input.Password
		role = input.Role
		zona = input.Zona

		if input.Image != "" {
			imageBase64 = input.Image
			imageUpdated = true
		}
	}

	updates := map[string]interface{}{}

	if username != "" {
		updates["username"] = username
	}
	if role != "" {
		updates["role"] = role
	}
	if zona != "" {
		updates["zona"] = zona
	}
	if password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error al encriptar la contraseña", http.StatusInternalServerError)
			return
		}
		updates["password"] = string(hashed)
	}
	if imageUpdated {
        decodedImage, err := base64.StdEncoding.DecodeString(imageBase64)
        if err != nil {
        	http.Error(w, "Error al decodificar imagen", http.StatusBadRequest)
        	return
        }
        updates["image"] = decodedImage
	}

	if len(updates) > 0 {
		if err := db.DB.Model(&user).Updates(updates).Error; err != nil {
			http.Error(w, "Error al actualizar usuario: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

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
