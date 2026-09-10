package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	userv1 "github.com/UmedjonQurbonov/cinema-libs/gen/go/user/v1"
	"github.com/UmedjonQurbonov/cinema-libs/pkg/postgres"
	server "github.com/UmedjonQurbonov/cinema-user-service/internal/delivery/http/v1"
	"github.com/UmedjonQurbonov/cinema-user-service/internal/repository"
	redisRepo "github.com/UmedjonQurbonov/cinema-user-service/internal/repository/redis"
	"github.com/UmedjonQurbonov/cinema-user-service/internal/service"
	"github.com/UmedjonQurbonov/cinema-user-service/pkg/mailer"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	if err := godotenv.Load(); err != nil {
		logger.Warn("no .env file found, using system environment variables")
	}

	dbPool, err := postgres.NewPgxPool(ctx, logger) 
	if err != nil {
		logger.Fatal("failed to initialize postgres pool", zap.Error(err))
	}

	defer dbPool.Close()

	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})

	ctxPing, cancelPing := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelPing()
	if err := rdb.Ping(ctxPing).Err(); err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	tempStore := redisRepo.NewTempUserRepository(rdb)

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := getEnv("SMTP_PORT", "587")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	

	if smtpUser == "" || smtpPassword == "" {
		logger.Fatal("SMTP_USER and SMTP_PASSWORD must be set in .env")
	}

	mailSender := mailer.NewSMTPSender(smtpHost, smtpPort, smtpUser, smtpPassword)

	logger.Info("application initialized successfully")

	lis, err := net.Listen("tcp", ":50051")

	if err != nil {
		log.Fatal("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer()

	userRepository := repository.NewUserRepo(dbPool)
	userService := service.NewServiceUser(userRepository, tempStore, mailSender)
	userServer := server.NewServerUser(userService, logger)

	userv1.RegisterUserServiceServer(grpcServer, userServer)

	reflection.Register(grpcServer)

	go func() {
		log.Printf("server listening at %v", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("failed to serve: %w", err)
		}
	} ()

	sigChan := make(chan os.Signal, 1) 
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<- sigChan
	ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	log.Println("shutting down server...")
	grpcServer.GracefulStop()
	log.Println("server stopped")
}