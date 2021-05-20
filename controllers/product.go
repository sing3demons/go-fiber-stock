package controllers

import (
	"app/models"
	"mime/multipart"
	"os"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type productForm struct {
	Name  string                `form:"name" validate:"required"`
	Stock int                   `form:"stock" validate:"required"`
	Price int                   `form:"price" validate:"required"`
	Image *multipart.FileHeader `form:"image" validate:"required"`
}

type updateProductForm struct {
	Name  string                `form:"name"`
	Stock int                   `form:"stock"`
	Price int                   `form:"price"`
	Image *multipart.FileHeader `form:"image"`
}

type productRespons struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Stock int    `json:"stock"`
	Price int    `json:"price"`
	Image string `json:"image"`
}

type Product struct {
	DB *gorm.DB
}

func (p *Product) FindAll(ctx *fiber.Ctx) error {
	var products []models.Product

	p.DB.Order("id desc").Find(&products)

	serializedProducts := []productRespons{}
	copier.Copy(&serializedProducts, &products)
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"products": serializedProducts})
}

func (p *Product) FindOne(ctx *fiber.Ctx) error {
	product, err := p.findProductByID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	serializedProduct := productRespons{}
	copier.Copy(&serializedProduct, &product)
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"product": serializedProduct})
}

func (p *Product) Create(ctx *fiber.Ctx) error {
	var form productForm
	if err := ctx.BodyParser(&form); err != nil {
		ctx.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": err.Error()})
		return err
	}

	var product models.Product
	copier.Copy(&product, &form)
	p.DB.Create(&product)

	p.setProductImage(ctx, &product)

	var serializedProduct productRespons
	copier.Copy(&serializedProduct, &product)

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"product": serializedProduct})
}

func (p *Product) Update(ctx *fiber.Ctx) error {
	var form updateProductForm
	if err := ctx.BodyParser(&form); err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	product, err := p.findProductByID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	copier.Copy(&product, &form)

	if err := p.DB.Save(&product).Error; err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": err.Error()})
	}

	p.setProductImage(ctx, product)

	return ctx.SendStatus(fiber.StatusOK)
}

//Delete - delete product
func (p *Product) Delete(ctx *fiber.Ctx) error {
	product, err := p.findProductByID(ctx)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	p.DB.Unscoped().Delete(&product)
	p.removeImageProduct(ctx, product)
	return ctx.SendStatus(fiber.StatusNoContent)
}

func (p *Product) findProductByID(ctx *fiber.Ctx) (*models.Product, error) {
	var product models.Product
	id := ctx.Params("id")

	if err := p.DB.First(&product, id).Error; err != nil {
		return nil, err
	}

	return &product, nil

}

func (p *Product) setProductImage(ctx *fiber.Ctx, product *models.Product) error {
	file, err := ctx.FormFile("image")
	if err != nil || file == nil {
		return err
	}

	p.removeImageProduct(ctx, product)
	path := "uploads/products/" + strconv.Itoa(int(product.ID))
	os.MkdirAll(path, 0755)

	filename := path + "/" + file.Filename
	if err := ctx.SaveFile(file, filename); err != nil {
		return err
	}

	product.Image = os.Getenv("HOST") + "/" + filename

	if err := p.DB.Save(product).Error; err != nil {
		return err
	}

	return nil

}

func (p *Product) removeImageProduct(ctx *fiber.Ctx, product *models.Product) error {
	if product.Image != "" {
		product.Image = strings.Replace(product.Image, os.Getenv("HOST"), "", 1)
		pwd, _ := os.Getwd()
		os.Remove(pwd + product.Image)
	}
	return nil
}
