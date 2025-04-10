package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/JMURv/avito-spring/internal/config"
	dto "github.com/JMURv/avito-spring/internal/dto/gen"
	md "github.com/JMURv/avito-spring/internal/models"
	"github.com/JMURv/avito-spring/internal/repo"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"time"
)

type Repository struct {
	conn *sqlx.DB
}

func New(conf config.Config) *Repository {
	conn, err := sqlx.Open(
		"pgx", fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=disable",
			conf.DB.User,
			conf.DB.Password,
			conf.DB.Host,
			conf.DB.Port,
			conf.DB.Database,
		),
	)
	if err != nil {
		zap.L().Fatal("Failed to connect to the database", zap.Error(err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = conn.PingContext(ctx); err != nil {
		zap.L().Fatal("Failed to ping the database", zap.Error(err))
	}

	if err = applyMigrations(conn.DB, conf); err != nil {
		zap.L().Fatal("Failed to apply migrations", zap.Error(err))
	}

	return &Repository{conn: conn}
}

func (r *Repository) Close() error {
	return r.conn.Close()
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*md.User, error) {
	var res md.User
	err := r.conn.GetContext(ctx, &res, getUserByEmail, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repo.ErrNotFound
		}
		return nil, err
	}
	return &res, nil
}

func (r *Repository) CreateUser(ctx context.Context, req *dto.RegisterPostReq) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.conn.QueryRowContext(
		ctx, createUser,
		req.Email,
		req.Password,
		req.Role,
	).Scan(&id)

	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (r *Repository) CreatePVZ(ctx context.Context, req *dto.PVZ) (uuid.UUID, time.Time, error) {
	var id uuid.UUID
	var createdAt time.Time
	err := r.conn.QueryRowContext(ctx, createPVZ, req.City).Scan(&id, &createdAt)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "22P02" {
				return uuid.Nil, time.Time{}, repo.ErrCityIsNotValid
			}
		}
		return uuid.Nil, time.Time{}, err
	}
	return id, createdAt, nil
}

func (r *Repository) GetPVZ(ctx context.Context, page, limit int64, startDate, endDate time.Time) ([]*dto.PvzGetOKItem, error) {
	rows, err := r.conn.QueryxContext(ctx, getPVZ, startDate, endDate, limit, (page-1)*limit)
	if err != nil {
		return nil, err
	}

	defer func(rows *sqlx.Rows) {
		if err := rows.Close(); err != nil {
			zap.L().Error("Failed to close rows", zap.Error(err))
		}
	}(rows)

	pvzMap := make(map[string]*dto.PvzGetOKItem)
	receptionMap := make(map[string]map[string]int)
	for rows.Next() {
		var (
			pvzID           uuid.UUID
			pvzCity         string
			pvzCreatedAt    time.Time
			receptionID     uuid.UUID
			receptionDate   time.Time
			receptionStatus string
			productID       uuid.UUID
			productDate     time.Time
			productType     string
		)

		if err := rows.Scan(
			&pvzID,
			&pvzCity,
			&pvzCreatedAt,
			&receptionID,
			&receptionDate,
			&receptionStatus,
			&productID,
			&productDate,
			&productType,
		); err != nil {
			return nil, err
		}

		pvzKey := pvzID.String()
		if _, ok := pvzMap[pvzKey]; !ok {
			pvzMap[pvzKey] = &dto.PvzGetOKItem{
				Pvz: dto.OptPVZ{
					Set: true,
					Value: dto.PVZ{
						ID: dto.OptUUID{
							Set:   true,
							Value: pvzID,
						},
						City: dto.PVZCity(pvzCity),
						RegistrationDate: dto.OptDateTime{
							Set:   true,
							Value: pvzCreatedAt,
						},
					},
				},
				Receptions: make([]dto.PvzGetOKItemReceptionsItem, 0),
			}
			receptionMap[pvzKey] = make(map[string]int)
		}

		if receptionID == uuid.Nil {
			continue
		}

		receptionKey := receptionID.String()
		currPVZ := pvzMap[pvzKey]
		if idx, ok := receptionMap[pvzKey][receptionKey]; ok {
			currPVZ.Receptions[idx].Products = append(
				currPVZ.Receptions[idx].Products, dto.Product{
					ID: dto.OptUUID{
						Set:   true,
						Value: productID,
					},
					DateTime: dto.OptDateTime{
						Set:   true,
						Value: productDate,
					},
					Type:        dto.ProductType(productType),
					ReceptionId: receptionID,
				},
			)
		} else {
			newReception := dto.PvzGetOKItemReceptionsItem{
				Reception: dto.OptReception{
					Set: true,
					Value: dto.Reception{
						ID: dto.OptUUID{
							Set:   true,
							Value: receptionID,
						},
						DateTime: receptionDate,
						PvzId:    pvzID,
						Status:   dto.ReceptionStatus(receptionStatus),
					},
				},
				Products: []dto.Product{
					{
						ID: dto.OptUUID{
							Set:   true,
							Value: productID,
						},
						DateTime: dto.OptDateTime{
							Set:   true,
							Value: productDate,
						},
						Type:        dto.ProductType(productType),
						ReceptionId: receptionID,
					},
				},
			}
			currPVZ.Receptions = append(currPVZ.Receptions, newReception)
			receptionMap[pvzKey][receptionKey] = len(currPVZ.Receptions) - 1
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	result := make([]*dto.PvzGetOKItem, 0, len(pvzMap))
	for _, pvz := range pvzMap {
		result = append(result, pvz)
	}
	return result, nil
}

func (r *Repository) CloseLastReception(ctx context.Context, id uuid.UUID) (*dto.Reception, error) {
	tx, err := r.conn.BeginTxx(
		ctx, &sql.TxOptions{
			Isolation: sql.LevelRepeatableRead,
		},
	)
	if err != nil {
		return nil, err
	}

	defer func(tx *sqlx.Tx) {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			zap.L().Error("Failed to rollback transaction", zap.Error(err))
		}
	}(tx)

	var res md.Reception
	err = tx.GetContext(ctx, &res, findLastReceptionForUpdate, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repo.ErrReceptionAlreadyClosed
		}
		return nil, err
	}

	_, err = tx.ExecContext(ctx, closeReception, res.ID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &dto.Reception{
		ID: dto.OptUUID{
			Set:   true,
			Value: res.ID,
		},
		DateTime: res.DateTime,
		PvzId:    res.PVZID,
		Status:   dto.ReceptionStatus(res.Status),
	}, nil
}

func (r *Repository) DeleteLastProduct(ctx context.Context, id uuid.UUID) error {
	tx, err := r.conn.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer func(tx *sqlx.Tx) {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			zap.L().Error("Failed to rollback transaction", zap.Error(err))
		}
	}(tx)

	var reception md.Reception
	err = tx.GetContext(ctx, &reception, findLastReception, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return repo.ErrNoActiveReception
		}
		return err
	}

	res, err := tx.ExecContext(ctx, deleteLastProduct, reception.ID)
	if err != nil {
		return err
	}

	if aff, _ := res.RowsAffected(); aff == 0 {
		return repo.ErrNoItems
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *Repository) CreateReception(ctx context.Context, req *dto.ReceptionsPostReq) (*dto.Reception, error) {
	tx, err := r.conn.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func(tx *sqlx.Tx) {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			zap.L().Error("Failed to rollback transaction", zap.Error(err))
		}
	}(tx)

	var res md.Reception
	err = tx.GetContext(ctx, &res, findLastReceptionForUpdate, req.PvzId)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	} else {
		return nil, repo.ErrReceptionStillOpen
	}

	err = r.conn.GetContext(ctx, &res, createReception, req.PvzId)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &dto.Reception{
		ID: dto.OptUUID{
			Set:   true,
			Value: res.ID,
		},
		DateTime: res.DateTime,
		PvzId:    res.PVZID,
		Status:   dto.ReceptionStatus(res.Status),
	}, nil
}

func (r *Repository) AddItemToReception(ctx context.Context, req *dto.ProductsPostReq) (*dto.Product, error) {
	var reception md.Reception
	err := r.conn.GetContext(ctx, &reception, findLastReception, req.PvzId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repo.ErrNoActiveReception
		}
		return nil, err
	}

	var res md.Product
	err = r.conn.GetContext(ctx, &res, addItemToReception, reception.ID, req.Type)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "22P02" {
				return nil, repo.ErrTypeIsNotValid
			}
		}
		return nil, err
	}

	return &dto.Product{
		ID: dto.OptUUID{
			Set:   true,
			Value: res.ID,
		},
		DateTime: dto.OptDateTime{
			Set:   true,
			Value: res.DateTime,
		},
		Type:        dto.ProductType(res.Type),
		ReceptionId: res.ReceptionId,
	}, nil
}

func (r *Repository) GetPVZList(ctx context.Context) ([]*md.PVZ, error) {
	var res []*md.PVZ
	err := r.conn.SelectContext(ctx, &res, listPVZs)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return res, nil
		}
		return nil, err
	}

	return res, err
}
