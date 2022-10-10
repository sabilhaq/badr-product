package main

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	app := fiber.New()
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:3306)/badr"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&Product{}, &Attribute{})

	app.Get("/", func(c *fiber.Ctx) error {
		var results []Join
		db.Model(&Product{}).Select("products.id, attributes.name").Joins("left join attributes on attributes.product_id = products.id").Scan(&results)
		return c.JSON(&ResponseJSON{
			Data: results,
		})
	})

	app.Post("/", func(c *fiber.Ctx) error {
		body := new(Join)

		if err := c.BodyParser(body); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		product := &Product{}

		db.Create(product)
		db.Create(&Attribute{Name: body.Name, ProductID: product.ID})

		return c.JSON(&ResponseJSON{
			Data: product,
		})
	})

	app.Get("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		var results []Join
		db.Model(&Product{}).Where("products.id = ?", id).Select("products.id, attributes.name").Joins("left join attributes on attributes.product_id = products.id").Scan(&results)
		return c.JSON(&ResponseJSON{
			Data: results[0],
		})
	})

	app.Put("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		body := new(Join)
		if err := c.BodyParser(body); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		product := &Product{}
		productId, _ := strconv.ParseUint(id, 10, 64)
		db.Model(&Attribute{}).Where("product_id = ?", productId).Updates(&Attribute{Name: body.Name, ProductID: uint(productId)})
		return c.JSON(&ResponseJSON{
			Data: product,
		})
	})

	app.Delete("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		productId, _ := strconv.ParseUint(id, 10, 64)
		db.Where("id = ?", uint(productId)).Delete(&Product{})
		return c.JSON(&ResponseJSON{
			Data: nil,
		})
	})

	log.Fatal(app.Listen(":3000"))
}

type Join struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Product struct {
	gorm.Model
}

type Attribute struct {
	gorm.Model
	Name      string `json:"name"`
	ProductID uint   `json:"product_id"`
}

type ResponseJSON struct {
	Data interface{} `json:"data"`
}
