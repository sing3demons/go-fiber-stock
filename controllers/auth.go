package controllers

import (
	"app/models"
	"fmt"
	"mime/multipart"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/copier"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	jwt "github.com/form3tech-oss/jwt-go"
	jwtware "github.com/gofiber/jwt/v2"
)

type Auth struct {
	DB *gorm.DB
}

type CreateUser struct {
	Name            string `json:"name" validate:"required"`
	Email           string `json:"email" validate:"required"`
	Password        string `json:"password" validate:"required"`
	PasswordConfirm string `json:"password_confirm" validate:"required"`
}

type updateProfileForm struct {
	Name   string                `form:"name"`
	Email  string                `form:"email" `
	Avatar *multipart.FileHeader `form:"avatar"`
}

type userResponse struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Avatar string `json:"avatar"`
	Role   string `json:"role"`
}
type loginUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Claims struct {
	jwt.StandardClaims
}

func (a *Auth) Register(ctx *fiber.Ctx) error {
	var user models.User
	var form CreateUser
	if err := ctx.BodyParser(&form); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if form.Password != form.PasswordConfirm {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Password do not macth"})
	}

	copier.Copy(&user, &form)
	user.Password = user.GenerateFromPassword()
	a.DB.Create(&user)

	var serializedUser userResponse
	copier.Copy(&serializedUser, &user)
	return ctx.Status(fiber.StatusCreated).JSON(serializedUser)
}

func (a *Auth) Login(ctx *fiber.Ctx) error {
	var form loginUser
	if err := ctx.BodyParser(&form); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	var user models.User

	if err := a.DB.Where("email = ?", form.Email).First(&user).Error; err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.Password)); err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})

	}

	var serializedUser userResponse
	copier.Copy(&serializedUser, &user)

	jwtToken := jwt.New(jwt.SigningMethodHS256)

	claims := jwtToken.Claims.(jwt.MapClaims)
	claims["id"] = user.ID
	// claims["user"] = serializedUser
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	token, err := jwtToken.SignedString([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}

	return ctx.JSON(fiber.Map{"token": token})

}

func Authenticate() fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey:   []byte(os.Getenv("SECRET_KEY")),
		ErrorHandler: jwtError,
		ContextKey:   "sub",
		// SuccessHandler: Authorize(),
	})
}

func jwtError(c *fiber.Ctx, err error) error {
	if err.Error() == "Missing or malformed JWT" {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"status": "error", "message": "Missing or malformed JWT", "data": nil})
	}
	return c.Status(fiber.StatusUnauthorized).
		JSON(fiber.Map{"status": "error", "message": "Invalid or expired JWT", "data": nil})
}

func (a *Auth) GetProfile(ctx *fiber.Ctx) error {

	contextKey := ctx.Locals("sub").(*jwt.Token)
	claims := contextKey.Claims.(jwt.MapClaims)
	id := claims["id"]

	var user models.User
	if err := a.DB.First(&user, id).Error; err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	var serializedUser userResponse
	copier.Copy(&serializedUser, &user)

	return ctx.JSON(fiber.Map{"user": serializedUser})
}

// func (a *Auth) Logout(ctx *fiber.Ctx) error {
// 	cookie := fiber.Cookie{
// 		Name:     "token",
// 		Value:    "",
// 		Expires:  time.Now().Add(-time.Hour),
// 		HTTPOnly: true,
// 	}
// 	ctx.Cookie(&cookie)
// 	return ctx.JSON(fiber.Map{"message": "success"})
// }

// func (a *Auth) UpdateImageProfile(ctx *fiber.Ctx) error {
// 	var form updateProfileForm
// 	if err := ctx.ShouldBind(&form); err != nil {
// 		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
// 		return
// 	}
// 	contextKey := ctx.Locals("sub").(*jwt.Token)
// 	claims := contextKey.Claims.(jwt.MapClaims)
// 	id := claims["id"]

// 	setUserImage(ctx, &user)

// 	var serializedUser userResponse
// 	copier.Copy(&serializedUser, &user)
// 	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"user": serializedUser})

// }

//UpdateProfile - PUT /api/v1/profile
// func (a *Auth) UpdateProfile(ctx *fiber.Ctx) {
// 	form := updateProfileForm{}
// 	if err := ctx.BodyParser(&form); err != nil {
// 		ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
// 		return
// 	}

// 	sub, _ := ctx.Get("sub")
// 	user := sub.(models.User)
// 	copier.Copy(&user, &form)

// 	if err := a.DB.Save(&user).Error; err != nil {
// 		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
// 		return
// 	}
// }

func Authorize() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		contextKey := ctx.Locals("sub").(*jwt.Token)
		claims := contextKey.Claims.(jwt.MapClaims)
		user := claims["user"]

		fmt.Print(user)

		// ctx.Next()
		return ctx.Next()
	}
}
