package service

import (
	"context"
	"errors"
	"testing"

	"gin-demo/internal/dto"
	"gin-demo/internal/model"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type fakeUserRepository struct {
	created   *model.User
	createErr error
}

func (r *fakeUserRepository) Create(_ context.Context, user *model.User) error {
	if r.createErr != nil {
		return r.createErr
	}
	r.created = user
	user.ID = 1
	return nil
}

func TestCreateUserNormalizesAndHashesPassword(t *testing.T) {
	repository := &fakeUserRepository{}
	service := NewUserService(repository)

	user, err := service.Create(context.Background(), dto.CreateUserRequest{
		Username: "  Hielo_19  ",
		Email:    "  USER@Example.COM  ",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if user.Username != "hielo_19" || user.Email != "user@example.com" {
		t.Fatalf("Create() normalized values = (%q, %q)", user.Username, user.Email)
	}
	if user.Role != "author" || user.Status != "active" {
		t.Fatalf("Create() defaults = (%q, %q)", user.Role, user.Status)
	}
	if user.PasswordHash == "password123" {
		t.Fatal("Create() stored the plaintext password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("password123")); err != nil {
		t.Fatalf("password hash does not match: %v", err)
	}
}

func TestCreateUserRejectsInvalidUsername(t *testing.T) {
	service := NewUserService(&fakeUserRepository{})

	_, err := service.Create(context.Background(), dto.CreateUserRequest{
		Username: "invalid name",
		Email:    "user@example.com",
		Password: "password123",
	})
	if !errors.Is(err, ErrInvalidUsername) {
		t.Fatalf("Create() error = %v, want ErrInvalidUsername", err)
	}
}

func TestCreateUserMapsDuplicateEmail(t *testing.T) {
	service := NewUserService(&fakeUserRepository{
		createErr: &pgconn.PgError{
			Code:           "23505",
			ConstraintName: "uk_users_email",
		},
	})

	_, err := service.Create(context.Background(), dto.CreateUserRequest{
		Username: "new_user",
		Email:    "used@example.com",
		Password: "password123",
	})
	if !errors.Is(err, ErrEmailAlreadyUsed) {
		t.Fatalf("Create() error = %v, want ErrEmailAlreadyUsed", err)
	}
}
