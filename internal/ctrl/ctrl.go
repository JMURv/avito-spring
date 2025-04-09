package ctrl

import (
	"context"
	"errors"
	"github.com/JMURv/avito-spring/internal/auth"
	"github.com/JMURv/avito-spring/internal/dto"
	md "github.com/JMURv/avito-spring/internal/models"
	metrics "github.com/JMURv/avito-spring/internal/observability/metrics/prometheus"
	"github.com/JMURv/avito-spring/internal/repo"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"time"
)

type AppRepo interface {
	GetUserByEmail(ctx context.Context, email string) (*md.User, error)
	CreateUser(ctx context.Context, req *dto.RegisterRequest) (uuid.UUID, error)
	CreatePVZ(ctx context.Context, req *dto.CreatePVZRequest) (uuid.UUID, time.Time, error)
	GetPVZ(ctx context.Context, page, limit int64, startDate, endDate time.Time) ([]*dto.GetPVZResponse, error)
	CloseLastReception(ctx context.Context, id uuid.UUID) (*md.Reception, error)
	DeleteLastProduct(ctx context.Context, id uuid.UUID) error
	CreateReception(ctx context.Context, req *dto.CreateReceptionRequest) (*dto.CreateReceptionResponse, error)
	AddItemToReception(ctx context.Context, req *dto.AddItemRequest) (*dto.AddItemResponse, error)

	GetPVZList(ctx context.Context) ([]*md.PVZ, error)
}

type AppCtrl interface {
	DummyLogin(ctx context.Context, req *dto.DummyLoginRequest) (*dto.DummyLoginResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error)
	GetPVZ(ctx context.Context, page, limit int64, startDate, endDate time.Time) ([]*dto.GetPVZResponse, error)
	CreatePVZ(ctx context.Context, req *dto.CreatePVZRequest) (*dto.CreatePVZResponse, error)
	CloseLastReception(ctx context.Context, id uuid.UUID) (*md.Reception, error)
	DeleteLastProduct(ctx context.Context, id uuid.UUID) error
	CreateReception(ctx context.Context, req *dto.CreateReceptionRequest) (*dto.CreateReceptionResponse, error)
	AddItemToReception(ctx context.Context, req *dto.AddItemRequest) (*dto.AddItemResponse, error)

	GetPVZList(ctx context.Context) ([]*md.PVZ, error)
}

type Controller struct {
	repo AppRepo
	au   auth.Core
}

func New(repo AppRepo, au auth.Core) *Controller {
	return &Controller{
		repo: repo,
		au:   au,
	}
}

func (c *Controller) DummyLogin(_ context.Context, req *dto.DummyLoginRequest) (*dto.DummyLoginResponse, error) {
	token, err := c.au.NewToken(uuid.New(), req.Role)
	if err != nil {
		return nil, err
	}

	return &dto.DummyLoginResponse{
		Token: token,
	}, nil
}

func (c *Controller) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	usr, err := c.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			zap.L().Debug("User not found", zap.String("email", req.Email))
			return nil, auth.ErrInvalidCredentials
		}
		zap.L().Error("Failed to get user by email", zap.Error(err))
		return nil, err
	}

	err = c.au.ComparePasswords([]byte(usr.Password), []byte(req.Password))
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			zap.L().Debug("Password mismatch", zap.String("email", req.Email))
			return nil, auth.ErrInvalidCredentials
		}
		zap.L().Error("Failed to compare passwords", zap.Error(err))
		return nil, err
	}

	token, err := c.au.NewToken(usr.ID, usr.Role)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		Token: token,
	}, nil
}

func (c *Controller) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	var err error
	var id uuid.UUID

	req.Password, err = c.au.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	id, err = c.repo.CreateUser(ctx, req)
	if err != nil {
		zap.L().Error("Failed to create user", zap.Error(err))
		return nil, err
	}

	return &dto.RegisterResponse{
		ID:    id,
		Email: req.Email,
		Role:  req.Role,
	}, nil
}

func (c *Controller) GetPVZ(ctx context.Context, page, limit int64, startDate, endDate time.Time) ([]*dto.GetPVZResponse, error) {
	res, err := c.repo.GetPVZ(ctx, page, limit, startDate, endDate)
	if err != nil {
		zap.L().Error("Failed to get PVZ", zap.Error(err))
		return nil, err
	}

	return res, nil
}

func (c *Controller) CreatePVZ(ctx context.Context, req *dto.CreatePVZRequest) (*dto.CreatePVZResponse, error) {
	id, createdAt, err := c.repo.CreatePVZ(ctx, req)
	if err != nil {
		if errors.Is(err, repo.ErrCityIsNotValid) {
			return nil, ErrCityIsNotValid
		}
		zap.L().Error("Failed to create PVZ", zap.Error(err))
		return nil, err
	}

	metrics.CreatedPVZ.Inc()
	return &dto.CreatePVZResponse{
		ID:               id,
		RegistrationDate: createdAt,
		City:             req.City,
	}, nil
}

func (c *Controller) CloseLastReception(ctx context.Context, id uuid.UUID) (*md.Reception, error) {
	res, err := c.repo.CloseLastReception(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrReceptionAlreadyClosed) {
			zap.L().Debug("Reception already closed", zap.String("id", id.String()))
			return nil, ErrReceptionAlreadyClosed
		}
		zap.L().Error("Failed to close last reception", zap.String("id", id.String()), zap.Error(err))
		return nil, err
	}

	return res, nil
}

func (c *Controller) DeleteLastProduct(ctx context.Context, id uuid.UUID) error {
	err := c.repo.DeleteLastProduct(ctx, id)
	if err != nil {
		if errors.Is(err, repo.ErrNoActiveReception) {
			zap.L().Debug("No active reception", zap.String("id", id.String()))
			return ErrNoActiveReception
		}
		if errors.Is(err, repo.ErrNoItems) {
			zap.L().Debug("No items for deletion", zap.String("id", id.String()))
			return ErrNoItems
		}
		zap.L().Error("Failed to delete last product ", zap.String("id", id.String()), zap.Error(err))
		return err
	}

	return nil
}

func (c *Controller) CreateReception(ctx context.Context, req *dto.CreateReceptionRequest) (*dto.CreateReceptionResponse, error) {
	res, err := c.repo.CreateReception(ctx, req)
	if err != nil {
		if errors.Is(err, repo.ErrReceptionStillOpen) {
			zap.L().Debug("Reception still open", zap.String("uid", req.PVZID.String()))
			return nil, ErrReceptionStillOpen
		}

		zap.L().Error("Failed to create reception", zap.String("uid", req.PVZID.String()), zap.Error(err))
		return nil, err
	}

	metrics.CreatedOrderReceipts.Inc()
	return res, nil
}

func (c *Controller) AddItemToReception(ctx context.Context, req *dto.AddItemRequest) (*dto.AddItemResponse, error) {
	res, err := c.repo.AddItemToReception(ctx, req)
	if err != nil {
		if errors.Is(err, repo.ErrNoActiveReception) {
			zap.L().Debug(
				"No active reception",
				zap.String("uid", req.PVZID.String()),
				zap.String("type", req.Type),
			)
			return nil, ErrNoActiveReception
		}

		if errors.Is(err, repo.ErrTypeIsNotValid) {
			zap.L().Debug(
				"Type is not valid",
				zap.String("uid", req.PVZID.String()),
				zap.String("type", req.Type),
			)
			return nil, ErrTypeIsNotValid
		}

		zap.L().Error(
			"Failed to create reception",
			zap.String("uid", req.PVZID.String()),
			zap.String("type", req.Type),
			zap.Error(err),
		)
		return nil, err
	}

	metrics.AddedProducts.Inc()
	return res, nil
}

func (c *Controller) GetPVZList(ctx context.Context) ([]*md.PVZ, error) {
	res, err := c.repo.GetPVZList(ctx)
	if err != nil {
		zap.L().Error("Failed to get pvzs list", zap.Error(err))
		return nil, err
	}

	return res, nil
}
