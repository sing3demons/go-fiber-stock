package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name     string `gorm:"not null"`
	Email    string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
	Avatar   string
	Role     string `gorm:"default:'Member'; not null"`
}

func (u *User) GenerateFromPassword() string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(u.Password), 14)
	return string(hash)
}

//Promote - update user --> edidter
func (u *User) Promote() {
	u.Role = "Editor"
}

//Demote - Change user --> edidter
func (u *User) Demote() {
	u.Role = "Member"
}
