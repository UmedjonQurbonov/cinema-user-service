package v1

import (
	"context"
	"errors"

	userv1 "github.com/UmedjonQurbonov/cinema-libs/gen/go/user/v1"
	"github.com/UmedjonQurbonov/cinema-user-service/internal/domain"
	"github.com/UmedjonQurbonov/cinema-user-service/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserService interface {
    SendVerificationCode(ctx context.Context, user *domain.User) error
    VerifyAndRegister(ctx context.Context, email, inputCode string) (int, error)
}


type UserServer struct {
    userv1.UnimplementedUserServiceServer
	service UserService
	log *zap.Logger
}

func NewServerUser(service UserService, log *zap.Logger) *UserServer {
	return &UserServer{
		service: service,
		log: log,
	}
}

// func (s *UserServer) Register(ctx context.Context, req *userv1.RegisterUserRequest) (*userv1.RegisterUserResponse, error) {
// 	user := &domain.User{
// 		Email: req.GetEmail(),
// 		Password_hash: req.GetPassword(),
// 		Full_name: req.GetFullName(),
// 	}

// 	id, err := s.service.Register(ctx, user)

// 	if err != nil {
// 		s.log.Error("failed to created user", 
// 		zap.String("email", req.GetEmail()),
// 		zap.Error(err),
// 	)
// 		return nil, status.Error(codes.Internal, "failed to register user")
// 	}

// 	s.log.Info("user created successfully",
// 		zap.Int64("id", int64(id)),
// 		zap.String("email", req.GetEmail()),
// 	)

// 	return &userv1.RegisterUserResponse{
// 		Id: int64(id),
// 	}, nil
// }

func (s *UserServer) SendVerificationCode(ctx context.Context, req *userv1.SendVerificationCodeRequest) (*userv1.SendVerificationCodeResponse, error) {
	reqUser := &domain.User{
		Full_name:     req.GetFullName(),
		Email:         req.GetEmail(),
		Password_hash: req.GetPassword(),
	}

	err := s.service.SendVerificationCode(ctx, reqUser)
	if err != nil {
		s.log.Error("failed to send verification code",
			zap.String("email", req.GetEmail()),
			zap.Error(err),
		)

		if errors.Is(err, service.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		if errors.Is(err, service.ErrInvalidData) {
			return nil, status.Error(codes.InvalidArgument, "invalid input data")
		}

		return nil, status.Error(codes.Internal, "failed to send verification code")
	}

	s.log.Info("verification code sent successfully",
		zap.String("email", req.GetEmail()),
	)

	return &userv1.SendVerificationCodeResponse{
		Message: "Verification code sent",
	}, nil
}

func (s *UserServer) VerifyAndRegister(ctx context.Context, req *userv1.VerifyAndRegisterRequest) (*userv1.VerifyAndRegisterResponse, error) {
	id, err := s.service.VerifyAndRegister(ctx, req.GetEmail(), req.GetCode())
	if err != nil {
		s.log.Error("failed to verify and register user",
			zap.String("email", req.GetEmail()),
			zap.Error(err),
		)

		if errors.Is(err, service.ErrCodeNotFound) || errors.Is(err, service.ErrInvalidCode) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return nil, status.Error(codes.Internal, "failed to register user")
	}

	s.log.Info("user successfully registered",
		zap.Int64("id", int64(id)),
		zap.String("email", req.GetEmail()),
	)

	return &userv1.VerifyAndRegisterResponse{
		Id: int64(id),
	}, nil
}