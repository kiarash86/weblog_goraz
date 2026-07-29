package main

import (
	"log"
	"net/http"
	"weblog/internal/config"
	"weblog/internal/db"
	"weblog/internal/handlers"
	"weblog/internal/middlewares"
	"weblog/internal/migration"
	"weblog/internal/repository"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {

	var cfg *config.Config
	err := godotenv.Load()
	if err != nil {
		log.Print("cant load env and use system\n")
	}
	cfg, err = config.Load()
	if err != nil {
		log.Fatalf("cant load config: %v", err)
	}

	pool := db.Connect(cfg)
	defer pool.Close()

	err = migration.Run(pool)
	if err != nil {
		log.Fatalf("cant run migrations: %v", err)
	}

	userRepo := repository.NewUserRepository(pool)
	boardRepo := repository.NewBoardRepository(pool)
	commentRepo := repository.NewCommentRepository(pool)
	boardShareRepo := repository.NewBoardShareRepository(pool)
	_ = userRepo
	_ = boardRepo
	_ = commentRepo
	_ = boardShareRepo

	authHandler := handlers.NewAuthHandler(userRepo, cfg.JWTKey)
	boardHandler := handlers.NewBoardHandler(boardRepo, boardShareRepo)
	boardShareHandler := handlers.NewBoardShareHandler(boardRepo, boardShareRepo, userRepo)
	commentHandler := handlers.NewCommentHandler(boardRepo, boardShareRepo, commentRepo)

	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"https://goraz-weblog.netlify.app/"},
		AllowMethods: []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	protected := e.Group("")
	protected.Use(middlewares.RequireAuth(cfg.JWTKey))
	e.GET("/health", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	e.POST("/signup", authHandler.SignUp)
	e.POST("/login", authHandler.Login)

	protected.DELETE("/weblog/:id", boardHandler.DeleteByID)
	protected.GET("/weblog", boardHandler.Feed)
	protected.GET("/weblog/:id", boardHandler.GetByID)
	protected.POST("/weblog", boardHandler.Create)
	protected.POST("/weblog/:id/share", boardShareHandler.Add)
	protected.GET("/weblog/:id/comment", commentHandler.ListCommentsOfBoard)
	protected.POST("/weblog/:id/comment", commentHandler.Create)
	protected.DELETE("/weblog/:id/comment/:commentId", commentHandler.DeleteByID)
	log.Fatal(e.Start(":" + cfg.Port))

}
