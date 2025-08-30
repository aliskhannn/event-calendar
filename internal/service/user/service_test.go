//go:build unit
// +build unit

package user

import (
	"context"
	mocksuserrepo "github.com/aliskhannn/calendar-service/internal/mocks/service/user"
	userrepo "github.com/aliskhannn/calendar-service/internal/repository/user"
	"testing"
	"time"

	"github.com/aliskhannn/calendar-service/internal/config"
	"github.com/aliskhannn/calendar-service/internal/model"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestCreateUser_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocksuserrepo.NewMockuserRepository(ctrl)
	svc := New(mockRepo, &config.Config{})

	ctx := context.Background()
	testUser := model.User{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
	}

	mockRepo.EXPECT().GetUserByEmail(ctx, testUser.Email).Return(nil, userrepo.ErrUserNotFound)
	mockRepo.EXPECT().CreateUser(ctx, gomock.Any()).Return(uuid.New(), nil)

	id, err := svc.Create(ctx, testUser)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, id)
}

func TestCreateUser_AlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocksuserrepo.NewMockuserRepository(ctrl)
	svc := New(mockRepo, &config.Config{})

	ctx := context.Background()
	testUser := model.User{
		Email: "john@example.com",
	}

	mockRepo.EXPECT().GetUserByEmail(ctx, testUser.Email).Return(&model.User{Email: testUser.Email}, nil)

	id, err := svc.Create(ctx, testUser)
	require.ErrorIs(t, err, ErrUserAlreadyExists)
	require.Equal(t, uuid.Nil, id)
}

func TestGetByEmail_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocksuserrepo.NewMockuserRepository(ctrl)
	cfg := &config.Config{JWT: config.JWT{Secret: "secret", TTL: time.Hour}}
	svc := New(mockRepo, cfg)

	ctx := context.Background()
	password := "password123"

	hash, _ := hashPassword(password)
	mockRepo.EXPECT().GetUserByEmail(ctx, "john@example.com").Return(&model.User{
		ID:       uuid.New(),
		Name:     "John",
		Email:    "john@example.com",
		Password: hash,
	}, nil)

	token, err := svc.GetByEmail(ctx, "john@example.com", password)
	require.NoError(t, err)
	require.NotEmpty(t, token)
}

func TestGetByEmail_InvalidCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocksuserrepo.NewMockuserRepository(ctrl)
	svc := New(mockRepo, &config.Config{})

	ctx := context.Background()
	password := "password123"

	// Пользователь не найден
	mockRepo.EXPECT().GetUserByEmail(ctx, "unknown@example.com").Return(nil, userrepo.ErrUserNotFound)

	_, err := svc.GetByEmail(ctx, "unknown@example.com", password)
	require.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestGetByEmail_WrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocksuserrepo.NewMockuserRepository(ctrl)
	svc := New(mockRepo, &config.Config{})

	ctx := context.Background()

	hash, _ := hashPassword("correctpass")

	mockRepo.EXPECT().GetUserByEmail(ctx, "john@example.com").Return(&model.User{
		ID:       uuid.New(),
		Name:     "John",
		Email:    "john@example.com",
		Password: hash,
	}, nil)

	_, err := svc.GetByEmail(ctx, "john@example.com", "wrongpass")
	require.ErrorIs(t, err, ErrInvalidCredentials)
}
