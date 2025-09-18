package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"be-kbtg/docs"
	"be-kbtg/models"
)

const jwtSecret = "dev-secret-change" // for demo only

func main() {
	// Init DB
	db, err := gorm.Open(sqlite.Open("app.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatal(err)
	}

	app := fiber.New()

	// Swagger static spec + UI
	app.Get("/swagger/doc.json", func(c *fiber.Ctx) error {
		c.Type("json")
		return c.SendString(docs.Spec)
	})
	app.Get("/swagger/*", swagger.New(swagger.Config{URL: "http://localhost:3000/swagger/doc.json"}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "hello world"})
	})

	// Auth group
	auth := app.Group("/auth")

	auth.Post("/register", func(c *fiber.Ctx) error {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
		}
		if body.Email == "" || body.Password == "" {
			return c.Status(400).JSON(fiber.Map{"error": "email & password required"})
		}
		if len(body.Password) < 6 {
			return c.Status(400).JSON(fiber.Map{"error": "password too short"})
		}
		// Check existing
		var cnt int64
		db.Model(&models.User{}).Where("email = ?", body.Email).Count(&cnt)
		if cnt > 0 {
			return c.Status(409).JSON(fiber.Map{"error": "email exists"})
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "hash error"})
		}
		usr := models.User{Email: body.Email, PasswordHash: string(hash)}
		if err := db.Create(&usr).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db error"})
		}
		return c.Status(201).JSON(fiber.Map{"id": usr.ID, "email": usr.Email, "created_at": usr.CreatedAt})
	})

	auth.Post("/login", func(c *fiber.Ctx) error {
		var body struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
		}
		var usr models.User
		if err := db.Where("email = ?", body.Email).First(&usr).Error; err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "invalid credentials"})
		}
		if bcrypt.CompareHashAndPassword([]byte(usr.PasswordHash), []byte(body.Password)) != nil {
			return c.Status(401).JSON(fiber.Map{"error": "invalid credentials"})
		}
		// Create token
		claims := jwt.MapClaims{
			"sub":   usr.ID,
			"email": usr.Email,
			"exp":   time.Now().Add(time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "token error"})
		}
		return c.JSON(fiber.Map{"token": signed})
	})

	if err := app.Listen(":3000"); err != nil {
		log.Fatal(err)
	}
}
