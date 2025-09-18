package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"be-kbtg/auth"
	"be-kbtg/config"
	"be-kbtg/docs"
	"be-kbtg/models"
)

// Server holds application dependencies and HTTP router.
type Server struct {
	app  *fiber.App
	db   *gorm.DB
	auth auth.Manager
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run wires dependencies and starts the HTTP server.
func run() error {
	cfg := config.Load()

	db, err := initDB(cfg)
	if err != nil {
		return err
	}

	s := &Server{
		app: fiber.New(),
		db:  db,
		auth: auth.NewManager(
			cfg.JWTSecret,
			cfg.JWTExpiresIn,
		),
	}
	s.registerRoutes()

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("starting server on %s", addr)
	return s.app.Listen(addr)
}

// initDB opens the database connection and performs migrations.
func initDB(cfg config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.AutoMigrate(&models.User{}); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

// registerRoutes attaches all HTTP routes.
func (s *Server) registerRoutes() {
	s.registerSystemRoutes()
	s.registerAuthRoutes()
	s.registerProfileRoutes()
}

func (s *Server) registerSystemRoutes() {
	s.app.Get("/swagger/doc.json", func(c *fiber.Ctx) error {
		c.Type("json")
		return c.SendString(docs.Spec)
	})
	s.app.Get("/swagger/*", swagger.New(swagger.Config{URL: "/swagger/doc.json"}))

	s.app.Get("/", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"message": "ok"}) })
}

func (s *Server) registerAuthRoutes() {
	grp := s.app.Group("/auth")
	grp.Post("/register", s.handleRegister)
	grp.Post("/login", s.handleLogin)
}

func (s *Server) registerProfileRoutes() {
	authRequired := func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return respondError(c, fiber.StatusUnauthorized, "missing bearer token")
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := s.auth.Parse(tokenStr)
		if err != nil {
			return respondError(c, fiber.StatusUnauthorized, "invalid token")
		}
		idFloat, ok := claims["sub"].(float64)
		if !ok {
			return respondError(c, fiber.StatusUnauthorized, "invalid subject")
		}
		var u models.User
		if err := s.db.First(&u, uint(idFloat)).Error; err != nil {
			return respondError(c, fiber.StatusUnauthorized, "user not found")
		}
		c.Locals("user", &u)
		return c.Next()
	}

	me := s.app.Group("/me", authRequired)
	me.Get("/", s.handleGetMe)
	me.Put("/", s.handleUpdateMe)
}

// handleRegister creates a new user account.
func (s *Server) handleRegister(c *fiber.Ctx) error {
	var body models.RegisterRequest
	if err := c.BodyParser(&body); err != nil {
		return respondError(c, fiber.StatusBadRequest, "invalid body")
	}
	if err := body.Validate(); err != nil {
		return respondError(c, fiber.StatusBadRequest, err.Error())
	}
	var cnt int64
	s.db.Model(&models.User{}).Where("email = ?", body.Email).Count(&cnt)
	if cnt > 0 {
		return respondError(c, fiber.StatusConflict, "email exists")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		return respondError(c, fiber.StatusInternalServerError, "hash error")
	}
	usr := models.User{Email: body.Email, PasswordHash: string(hash)}
	if err := s.db.Create(&usr).Error; err != nil {
		return respondError(c, fiber.StatusInternalServerError, "db error")
	}
	resp := models.RegisterResponse{ID: usr.ID, Email: usr.Email, CreatedAt: usr.CreatedAt}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

// handleLogin authenticates a user and returns a token.
func (s *Server) handleLogin(c *fiber.Ctx) error {
	var body models.LoginRequest
	if err := c.BodyParser(&body); err != nil {
		return respondError(c, fiber.StatusBadRequest, "invalid body")
	}
	var usr models.User
	if err := s.db.Where("email = ?", body.Email).First(&usr).Error; err != nil {
		return respondError(c, fiber.StatusUnauthorized, "invalid credentials")
	}
	if bcrypt.CompareHashAndPassword([]byte(usr.PasswordHash), []byte(body.Password)) != nil {
		return respondError(c, fiber.StatusUnauthorized, "invalid credentials")
	}
	token, err := s.auth.Generate(usr.ID, usr.Email)
	if err != nil {
		return respondError(c, fiber.StatusInternalServerError, "token error")
	}
	return c.JSON(models.TokenResponse{Token: token})
}

// handleGetMe returns current authenticated user.
func (s *Server) handleGetMe(c *fiber.Ctx) error {
	user, err := currentUser(c)
	if err != nil {
		return respondError(c, fiber.StatusUnauthorized, "user context missing")
	}
	return c.JSON(user)
}

// handleUpdateMe updates fields on the current user.
func (s *Server) handleUpdateMe(c *fiber.Ctx) error {
	user, err := currentUser(c)
	if err != nil {
		return respondError(c, fiber.StatusUnauthorized, "user context missing")
	}
	var body models.UpdateProfileRequest
	if err := c.BodyParser(&body); err != nil {
		return respondError(c, fiber.StatusBadRequest, "invalid body")
	}
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
	if err := s.db.Save(user).Error; err != nil {
		return respondError(c, fiber.StatusInternalServerError, "db error")
	}
	return c.JSON(user)
}

// respondError provides a consistent error response body.
func respondError(c *fiber.Ctx, status int, msg string) error {
	return c.Status(status).JSON(fiber.Map{"error": msg})
}

// currentUser extracts the *models.User from context.
func currentUser(c *fiber.Ctx) (*models.User, error) {
	v := c.Locals("user")
	if v == nil {
		return nil, errors.New("no user in context")
	}
	u, ok := v.(*models.User)
	if !ok {
		return nil, errors.New("invalid user type")
	}
	return u, nil
}
