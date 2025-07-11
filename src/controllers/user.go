package controllers

import (
	"api3/db"
	"api3/src/models"
	"api3/src/utils"
	"encoding/json"
	"log"
	"net/http"
	"os"
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
	contentType := r.Header.Get("Content-Type")

	var username, password, role, zona, imagePath string

	if strings.HasPrefix(contentType, "multipart/form-data") {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, "No se pudo parsear el formulario: "+err.Error(), http.StatusBadRequest)
			return
		}

		username = r.FormValue("username")
		password = r.FormValue("password")
		role = r.FormValue("role")
		zona = r.FormValue("zona")

		if role == "" {
			role = "user"
		}

		if username == "" || password == "" || zona == "" {
			http.Error(w, "Username, password y zona son obligatorios", http.StatusBadRequest)
			return
		}

		file, handler, err := r.FormFile("image")
		if err == nil {
			defer file.Close()
			os.MkdirAll("uploads", os.ModePerm)
			imagePath = "uploads/" + handler.Filename
			dst, err := utils.SaveFile(file, imagePath)
			if err != nil || dst == "" {
				http.Error(w, "No se pudo guardar la imagen: "+err.Error(), http.StatusInternalServerError)
				return
			}
		} else if err != http.ErrMissingFile {
			http.Error(w, "Error al procesar imagen: "+err.Error(), http.StatusBadRequest)
			return
		}

	} else {
		var input struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Role     string `json:"role"`
			Zona     string `json:"zona"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Error en el formato JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		username = input.Username
		password = input.Password
		role = input.Role
		zona = input.Zona

		if role == "" {
			role = "user"
		}

		if username == "" || password == "" || zona == "" {
			http.Error(w, "Username, password y zona son obligatorios", http.StatusBadRequest)
			return
		}
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error al encriptar la contraseña", http.StatusInternalServerError)
		return
	}

	user := models.User{
		Username: username,
		Password: string(hashedPwd),
		Role:     role,
		Zona:     zona,
		Image:    imagePath,
	}

	result := db.DB.Create(&user)
	if result.Error != nil {
		http.Error(w, "Error al registrar usuario: "+result.Error.Error(), http.StatusBadRequest)
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

	token, err := utils.GenerateToken(dbUser.ID, dbUser.Role, dbUser.Zona)
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

	var username, role, password string
	var imagePath string
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

		file, handler, err := r.FormFile("image")
		if err == nil {
			defer file.Close()

			// Crear carpeta uploads si no existe
			if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
				http.Error(w, "No se pudo crear carpeta de imágenes: "+err.Error(), http.StatusInternalServerError)
				return
			}

			newImagePath := "uploads/" + handler.Filename

			// Guardar archivo en disco
			savedPath, err := utils.SaveFile(file, newImagePath)
			if err != nil || savedPath == "" {
				http.Error(w, "No se pudo guardar la nueva imagen: "+err.Error(), http.StatusInternalServerError)
				return
			}

			// Eliminar imagen anterior si existía
			if user.Image != "" {
				if err := os.Remove(user.Image); err != nil && !os.IsNotExist(err) {
					log.Println("No se pudo eliminar imagen anterior:", err)
				}
			}

			imagePath = savedPath
			imageUpdated = true
		} else {
			if err != http.ErrMissingFile {
				http.Error(w, "Error al procesar imagen: "+err.Error(), http.StatusBadRequest)
				return
			}
		}
	} else {
		// JSON sin imagen
		var input struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Role     string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Error en el formato JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		username = input.Username
		password = input.Password
		role = input.Role
	}

	// Construir solo campos válidos
	updates := map[string]interface{}{}

	if username != "" {
		updates["username"] = username
	}
	if role != "" {
		updates["role"] = role
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
		updates["image"] = imagePath
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
