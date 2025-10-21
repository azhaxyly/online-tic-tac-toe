package services

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"tictactoe/internal/models"
	"tictactoe/internal/store"
	"unicode"

	"github.com/redis/go-redis/v9"
)

type SessionService struct {
	RDB   *redis.Client
	Store *store.UserStore
}

func NewSessionService(rdb *redis.Client, store *store.UserStore) *SessionService {
	return &SessionService{RDB: rdb, Store: store}
}

func (s *SessionService) Register(nickname, password string) (*models.User, error) {
	// 1. Валидация Nickname
	if len(strings.TrimSpace(nickname)) < 3 {
		return nil, fmt.Errorf("nickname must be at least 3 characters")
	}

	// 2. НОВАЯ ВАЛИДАЦИЯ ПАРОЛЯ
	if err := s.validatePassword(password); err != nil {
		return nil, err
	}

	// 3. Хэшируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 4. Создаем пользователя
	user, err := s.Store.CreateUser(nickname, string(hashedPassword))
	if err != nil {
		return nil, fmt.Errorf("failed to create user (nickname might be taken): %w", err)
	}

	return user, nil
}

func (s *SessionService) Login(nickname, password string) (*models.User, error) {
	// 1. Находим пользователя
	user, storedHash, err := s.Store.GetUserByNickname(nickname)
	if err != nil {
		return nil, fmt.Errorf("invalid nickname or password") // Не говорим, что именно не так
	}

	// 2. Сравниваем пароль
	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err != nil {
		// Ошибка (пароль не совпал)
		return nil, fmt.Errorf("invalid nickname or password")
	}

	// 3. Пароль верный
	return user, nil
}

func (s *SessionService) validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		// IsPunct - это знаки пунктуации, IsSymbol - другие символы
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}
