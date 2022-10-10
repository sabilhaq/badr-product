package main

import (
	"fmt"
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

	db.AutoMigrate(&Product{}, &Attribute{})

	app.Get("/", func(c *fiber.Ctx) error {
		var ams []Attribute
		_ = db.Model(&Product{}).Select("name, value, product_id").Joins("left join attributes on attributes.product_id = products.id").Scan(&ams)

		var idLast uint
		var attrsTemp []Attribute
		mapRes := map[uint][]Attribute{}
		for _, attribute := range ams {
			var id = attribute.ProductID

			if id == idLast {
				attrsTemp = append(attrsTemp, Attribute{
					Name:  attribute.Name,
					Value: attribute.Value,
				})
			} else {
				idLast = id
				attrsTemp = []Attribute{attribute}
			}
			mapRes[id] = attrsTemp
		}

		var resArr []ResponseBody
		for key, attributes := range mapRes {
			var attributesRes []AttributeRep

			for _, at := range attributes {
				attributesRes = append(attributesRes, AttributeRep{
					Name:  at.Name,
					Value: at.Value,
				})
			}

			mapAttribute := make(map[string]interface{})
			for _, am := range attributesRes {
				val, err := strconv.ParseInt(am.Value, 10, 64)
				if err != nil {
					mapAttribute[am.Name] = am.Value
				} else {
					mapAttribute[am.Name] = val
				}
			}

			resArr = append(resArr, ResponseBody{
				ID:        int(key),
				Attribute: mapAttribute,
			})

		}

		return c.JSON(&ResponseJSON{
			Data: resArr,
		})
	})

	app.Get("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		var ams []Attribute
		_ = db.Model(&Product{}).Select("name, value, product_id").Joins("left join attributes on attributes.product_id = products.id").Where("products.id = ?", id).Scan(&ams)

		var respBody ResponseBody
		mapAttribute := make(map[string]interface{})
		respBody.ID = int(ams[0].ProductID)
		for _, am := range ams {
			val, err := strconv.ParseInt(am.Value, 10, 64)
			if err != nil {
				mapAttribute[am.Name] = am.Value
			} else {
				mapAttribute[am.Name] = val
			}
		}
		respBody.Attribute = mapAttribute

		return c.JSON(&ResponseJSON{
			Data: respBody,
		})
	})

	app.Post("/", func(c *fiber.Ctx) error {
		request := new(RequestBody)
		if err := c.BodyParser(request); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		product := &Product{}
		db.Create(product)

		for key, val := range request.Attribute {
			db.Create(&Attribute{
				Name:      key,
				Value:     fmt.Sprintf("%v", val),
				ProductID: product.ID,
			})
		}

		return c.JSON(&ResponseJSON{
			Data: map[string]int{
				"id": int(product.ID),
			},
		})
	})

	app.Put("/:id", func(c *fiber.Ctx) error {
		productId, _ := strconv.ParseUint(c.Params("id"), 10, 64)

		request := new(RequestBody)
		if err := c.BodyParser(request); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		for key, val := range request.Attribute {
			db.Model(&Attribute{}).Where("product_id = ? and name = ?", productId, key).Update("value", val)
		}

		return c.JSON(&ResponseJSON{
			Data: map[string]uint64{
				"id": productId,
			},
		})
	})

	app.Delete("/:id", func(c *fiber.Ctx) error {
		productId, _ := strconv.ParseUint(c.Params("id"), 10, 64)
		db.Delete(&Product{}, productId)
		return c.JSON(&ResponseJSON{
			Data: nil,
		})
	})

	log.Fatal(app.Listen(":3000"))
}

type RequestBody struct {
	Attribute map[string]interface{} `json:"attribute"`
}

type Join struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ResponseBody struct {
	ID        int                    `json:"id"`
	Attribute map[string]interface{} `json:"attribute"`
}

type Product struct {
	gorm.Model
}

type Attribute struct {
	gorm.Model
	Name      string
	Value     string
	ProductID uint
}

type AttributeRep struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ResponseJSON struct {
	Data interface{} `json:"data"`
}
