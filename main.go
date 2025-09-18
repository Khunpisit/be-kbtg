package main

import (
	"errors"
	"log"
	"os"
	"strings"
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
	// Init DB (allow override by env DB_PATH)
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "app.db"
	}
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
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

	// --- Auth middleware and profile endpoints ---
	authRequired := func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(401).JSON(fiber.Map{"error": "missing bearer token"})
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		tkn, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !tkn.Valid {
			return c.Status(401).JSON(fiber.Map{"error": "invalid token"})
		}
		claims, ok := tkn.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "invalid token claims"})
		}
		idFloat, ok := claims["sub"].(float64)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "invalid subject"})
		}
		var u models.User
		if err := db.First(&u, uint(idFloat)).Error; err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "user not found"})
		}
		c.Locals("user", &u)
		return c.Next()
	}

	me := app.Group("/me", authRequired)
	me.Get("/", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		return c.JSON(user)
	})
	me.Put("/", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*models.User)
		var body struct {
			FirstName   *string `json:"first_name"`
			LastName    *string `json:"last_name"`
			DisplayName *string `json:"display_name"`
			Phone       *string `json:"phone"`
			AvatarURL   *string `json:"avatar_url"`
			Bio         *string `json:"bio"`
		}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
		}
		// Update only provided fields
		if body.FirstName != nil {
			user.FirstName = strings.TrimSpace(*body.FirstName)
		}
		if body.LastName != nil {
			user.LastName = strings.TrimSpace(*body.LastName)
		}
		if body.DisplayName != nil {
			user.DisplayName = strings.TrimSpace(*body.DisplayName)
		}
		if body.Phone != nil {
			user.Phone = strings.TrimSpace(*body.Phone)
		}
		if body.AvatarURL != nil {
			user.AvatarURL = strings.TrimSpace(*body.AvatarURL)
		}
		if body.Bio != nil {
			user.Bio = strings.TrimSpace(*body.Bio)
		}
		if user.DisplayName == "" && (user.FirstName != "" || user.LastName != "") {
			user.DisplayName = strings.TrimSpace(strings.Trim(user.FirstName+" "+user.LastName, " "))
		}
		if err := db.Save(user).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "db error"})
		}
		return c.JSON(user)
	})
	// --- end profile endpoints ---

	if err := app.Listen(":3000"); err != nil {
		log.Fatal(err)
	}
}
