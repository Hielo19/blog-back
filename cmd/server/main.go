package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gin-demo/internal/config"
	"gin-demo/internal/database"
	"gin-demo/internal/handler"
	"gin-demo/internal/repository"
	"gin-demo/internal/router"
	"gin-demo/internal/service"
)

const shutdownTimeout = 10 * time.Second

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	configuration, err := config.Load(config.DefaultPath)
	if err != nil {
		return fmt.Errorf("加载配置: %w", err)
	}

	db, err := database.Connect(context.Background(), configuration.DatabaseURL())
	if err != nil {
		return fmt.Errorf("初始化数据库: %w", err)
	}
	defer db.Close()
	log.Print("connected to PostgreSQL")

	userRepository := repository.NewPostgresUserRepository(db)
	userService := service.NewUserService(userRepository)
	userHandler := handler.NewUserHandler(userService)
	categoryRepository := repository.NewPostgresCategoryRepository(db)
	categoryService := service.NewCategoryService(categoryRepository)
	categoryHandler := handler.NewCategoryHandler(categoryService)
	postRepository := repository.NewPostgresPostRepository(db)
	postService := service.NewPostService(postRepository, categoryRepository)
	postHandler := handler.NewPostHandler(postService)

	server := &http.Server{
		Addr:              configuration.Address(),
		Handler:           router.New(postHandler, userHandler, categoryHandler),
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverError := make(chan error, 1)
	go func() {
		log.Printf("listening on http://localhost%s", configuration.Address())
		serverError <- server.ListenAndServe()
	}()

	signalContext, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	select {
	case err := <-serverError:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("HTTP 服务退出: %w", err)
	case <-signalContext.Done():
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownContext); err != nil {
		return fmt.Errorf("关闭 HTTP 服务: %w", err)
	}

	return nil
}
