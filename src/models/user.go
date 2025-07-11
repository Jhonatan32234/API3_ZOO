package models

type User struct {
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Username string `gorm:"unique;not null" json:"username"`
	Password string `gorm:"not null" json:"-"`
	Role     string `gorm:"default:user;not null" json:"role"`
	Zona     string `gorm:"not null" json:"zona"`
	Image    string `gorm:"type:text" json:"image"`
}
