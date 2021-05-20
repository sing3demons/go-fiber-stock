package routes

import (
	"app/config"
	"app/controllers"

	"github.com/gofiber/fiber/v2"
)

func Serve(app *fiber.App) {
	db := config.GetDB()

	v1 := app.Group("api/v1")

	authenticate := controllers.Authenticate()

	authController := controllers.Auth{DB: db}
	authGroup := v1.Group("/auth")
	{
		authGroup.Post("/register", authController.Register)
		authGroup.Post("/login", authController.Login)
		authGroup.Get("/profile", authenticate, authController.GetProfile)
		// authGroup.Patch("/profile", authenticate, authController.UpdateImageProfile)
	}

	productController := controllers.Product{DB: db}
	productsGroup := v1.Group("/products")
	productsGroup.Get("", productController.FindAll)
	productsGroup.Get("/:id", productController.FindOne)
	// productsGroup.Use(authenticate)
	{

		productsGroup.Post("", productController.Create)
		productsGroup.Put("/:id", productController.Update)
		productsGroup.Delete("/:id", productController.Delete)
	}

}
