package usecase

import (
	"Merch_store-Avito_test_task/internal/models"
	"Merch_store-Avito_test_task/internal/pkg/auth"
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
)

type AuthUsecaseImpl struct {
	repo auth.AuthRepository
}

func NewAuthUsecase(repo auth.AuthRepository) *AuthUsecaseImpl {
	return &AuthUsecaseImpl{repo}
}

func (uc *AuthUsecaseImpl) Login(ctx context.Context, username, password string) (models.User, error) {
	user, err := uc.repo.GetUser(ctx, username)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return models.User{}, fmt.Errorf("error hashing password: %w", err)
			}

			user := models.User{
				Username:     username,
				PasswordHash: string(hashedPassword),
			}

			userID, err := uc.repo.CreateUser(ctx, user)
			if err != nil {
				return models.User{}, fmt.Errorf("error creating user: %w", err)
			}

			user.ID = userID
			return user, nil
		}
		return models.User{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		log.Printf("Password mismatch: %v\n", err)
		return models.User{}, models.ErrMismatch
	}

	log.Printf("password match\n")
	return user, nil
}
