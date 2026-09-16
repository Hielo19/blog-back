package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"

	"gin-demo/internal/dto"
	"gin-demo/internal/model"
	"gin-demo/internal/repository"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidUsername     = errors.New("用户名只能包含字母、数字和下划线，长度为 3 到 50 个字符")
	ErrInvalidPassword     = errors.New("密码长度至少为 8 个字符，并且不能超过 72 字节")
	ErrUsernameAlreadyUsed = errors.New("用户名已被使用")
	ErrEmailAlreadyUsed    = errors.New("邮箱已被使用")
	ErrUserAlreadyExists   = errors.New("用户名或邮箱已被使用")
)

var usernamePattern = regexp.MustCompile(`^[a-z0-9_]{3,50}$`)

type UserService interface {
	Create(ctx context.Context, request dto.CreateUserRequest) (*model.User, error)
}

type userService struct {
	repository repository.UserRepository
}

func NewUserService(repository repository.UserRepository) UserService {
	return &userService{repository: repository}
}

func (s *userService) Create(
	ctx context.Context,
	request dto.CreateUserRequest,
) (*model.User, error) {
	request.Username = strings.ToLower(strings.TrimSpace(request.Username))
	request.Email = strings.ToLower(strings.TrimSpace(request.Email))
	if !usernamePattern.MatchString(request.Username) {
		return nil, ErrInvalidUsername
	}
	if utf8.RuneCountInString(request.Password) < 8 || len([]byte(request.Password)) > 72 {
		return nil, ErrInvalidPassword
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(request.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:     request.Username,
		Email:        request.Email,
		PasswordHash: string(passwordHash),
		DisplayName:  trimOptional(request.DisplayName),
		AvatarURL:    trimOptional(request.AvatarURL),
		Bio:          trimOptional(request.Bio),
		Role:         "author",
		Status:       "active",
	}

	if err := s.repository.Create(ctx, user); err != nil {
		return nil, translateUserDatabaseError(err)
	}
	return user, nil
}

func translateUserDatabaseError(err error) error {
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != "23505" {
		return err
	}

	switch postgresError.ConstraintName {
	case "uk_users_username":
		return ErrUsernameAlreadyUsed
	case "uk_users_email":
		return ErrEmailAlreadyUsed
	default:
		return ErrUserAlreadyExists
	}
}
