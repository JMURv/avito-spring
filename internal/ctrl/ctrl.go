package ctrl

import (
	"context"
	"errors"
	"github.com/JMURv/avito-spring/internal/auth"
	dto "github.com/JMURv/avito-spring/internal/dto/gen"
	md "github.com/JMURv/avito-spring/internal/models"
	metrics "github.com/JMURv/avito-spring/internal/observability/metrics/prometheus"
	"github.com/JMURv/avito-spring/internal/repo"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"time"
)

type AppRepo interface {
	GetUserByEmail(ctx context.Context, email string) (*md.User, error)
	CreateUser(ctx context.Context, req *dto.RegisterPostReq) (uuid.UUID, error)
	CreatePVZ(ctx context.Context, req *dto.PVZ) (uuid.UUID, time.Time, error)
	GetPVZ(ctx context.Context, page, limit int64, startDate, endDate time.Time) ([]*dto.PvzGetOKItem, error)
	CloseLastReception(ctx context.Context, id uuid.UUID) (*dto.Reception, error)
	DeleteLastProduct(ctx context.Context, id uuid.UUID) error
	CreateReception(ctx context.Context, req *dto.ReceptionsPostReq) (*dto.Reception, error)
	AddItemToReception(ctx context.Context, req *dto.ProductsPostReq) (*dto.Product, error)

	GetPVZList(ctx context.Context) ([]*md.PVZ, error)
}

type AppCtrl interface {
	DummyLogin(ctx context.Context, req *dto.DummyLoginPostReq) (dto.Token, error)
	Login(ctx context.Context, req *dto.LoginPostReq) (dto.Token, error)
	Register(ctx context.Context, req *dto.RegisterPostReq) (*dto.User, error)
	GetPVZ(ctx context.Context, page, limit int64, startDate, endDate time.Time) ([]*dto.PvzGetOKItem, error)
	CreatePVZ(ctx context.Context, req *dto.PVZ) (*dto.PVZ, error)
	CloseLastReception(ctx context.Context, id uuid.UUID) (*dto.Reception, error)
	DeleteLastProduct(ctx context.Context, id uuid.UUID) error
	CreateReception(ctx context.Context, req *dto.ReceptionsPostReq) (*dto.Reception, error)
	AddItemToReception(ctx context.Context, req *dto.ProductsPostReq) (*dto.Product, error)

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

func (c *Controller) DummyLogin(_ context.Context, req *dto.DummyLoginPostReq) (dto.Token, error) {
	token, err := c.au.NewToken(uuid.New(), string(req.Role))
	if err != nil {
		return "", err
	}

	return dto.Token(token), nil
}

func (c *Controller) Login(ctx context.Context, req *dto.LoginPostReq) (dto.Token, error) {
	usr, err := c.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			zap.L().Debug("User not found", zap.String("email", req.Email))
			return "", auth.ErrInvalidCredentials
		}
		zap.L().Error("Failed to get user by email", zap.Error(err))
		return "", err
	}

	err = c.au.ComparePasswords([]byte(usr.Password), []byte(req.Password))
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			zap.L().Debug("Password mismatch", zap.String("email", req.Email))
			return "", auth.ErrInvalidCredentials
		}
		zap.L().Error("Failed to compare passwords", zap.Error(err))
		return "", err
	}

	token, err := c.au.NewToken(usr.ID, usr.Role)
	if err != nil {
		return "", err
	}

	return dto.Token(token), nil
}

func (c *Controller) Register(ctx context.Context, req *dto.RegisterPostReq) (*dto.User, error) {
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

	return &dto.User{
		ID: dto.OptUUID{
			Value: id,
			Set:   true,
		},
		Email: req.Email,
		Role:  dto.UserRole(req.Role),
	}, nil
}

func (c *Controller) GetPVZ(ctx context.Context, page, limit int64, startDate, endDate time.Time) ([]*dto.PvzGetOKItem, error) {
	res, err := c.repo.GetPVZ(ctx, page, limit, startDate, endDate)
	if err != nil {
		zap.L().Error("Failed to get PVZ", zap.Error(err))
		return nil, err
	}

	return res, nil
}

func (c *Controller) CreatePVZ(ctx context.Context, req *dto.PVZ) (*dto.PVZ, error) {
	id, createdAt, err := c.repo.CreatePVZ(ctx, req)
	if err != nil {
		zap.L().Error("Failed to create PVZ", zap.Error(err))
		return nil, err
	}

	metrics.CreatedPVZ.Inc()
	return &dto.PVZ{
		ID: dto.OptUUID{
			Value: id,
			Set:   true,
		},
		RegistrationDate: dto.OptDateTime{
			Value: createdAt,
			Set:   true,
		},
		City: req.City,
	}, nil
}

func (c *Controller) CloseLastReception(ctx context.Context, id uuid.UUID) (*dto.Reception, error) {
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

func (c *Controller) CreateReception(ctx context.Context, req *dto.ReceptionsPostReq) (*dto.Reception, error) {
	res, err := c.repo.CreateReception(ctx, req)
	if err != nil {
		if errors.Is(err, repo.ErrReceptionStillOpen) {
			zap.L().Debug("Reception still open", zap.String("uid", req.PvzId.String()))
			return nil, ErrReceptionStillOpen
		}

		zap.L().Error("Failed to create reception", zap.String("uid", req.PvzId.String()), zap.Error(err))
		return nil, err
	}

	metrics.CreatedOrderReceipts.Inc()
	return res, nil
}

func (c *Controller) AddItemToReception(ctx context.Context, req *dto.ProductsPostReq) (*dto.Product, error) {
	res, err := c.repo.AddItemToReception(ctx, req)
	if err != nil {
		if errors.Is(err, repo.ErrNoActiveReception) {
			zap.L().Debug(
				"No active reception",
				zap.String("uid", req.PvzId.String()),
				zap.String("type", string(req.Type)),
			)
			return nil, ErrNoActiveReception
		}

		if errors.Is(err, repo.ErrTypeIsNotValid) {
			zap.L().Debug(
				"Type is not valid",
				zap.String("uid", req.PvzId.String()),
				zap.String("type", string(req.Type)),
			)
			return nil, ErrTypeIsNotValid
		}

		zap.L().Error(
			"Failed to create reception",
			zap.String("uid", req.PvzId.String()),
			zap.String("type", string(req.Type)),
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
