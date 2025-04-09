package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/JMURv/avito-spring/internal/config"
	"github.com/JMURv/avito-spring/internal/dto"
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

func (r *Repository) CreateUser(ctx context.Context, req *dto.RegisterRequest) (uuid.UUID, error) {
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

func (r *Repository) CreatePVZ(ctx context.Context, req *dto.CreatePVZRequest) (uuid.UUID, time.Time, error) {
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

func (r *Repository) GetPVZ(ctx context.Context, page, limit int64, startDate, endDate time.Time) ([]*dto.GetPVZResponse, error) {
	rows, err := r.conn.QueryxContext(ctx, getPVZ, startDate, endDate, limit, (page-1)*limit)
	if err != nil {
		return nil, err
	}

	defer func(rows *sqlx.Rows) {
		if err := rows.Close(); err != nil {
			zap.L().Error("Failed to close rows", zap.Error(err))
		}
	}(rows)

	pvzMap := make(map[string]*dto.GetPVZResponse)
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
			pvzMap[pvzKey] = &dto.GetPVZResponse{
				PVZ: md.PVZ{
					ID:               pvzID,
					City:             pvzCity,
					RegistrationDate: pvzCreatedAt,
				},
				Receptions: make(
					[]struct {
						Reception md.Reception `json:"reception"`
						Products  []md.Product `json:"products"`
					}, 0,
				),
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
				currPVZ.Receptions[idx].Products, md.Product{
					ID:          productID,
					DateTime:    productDate,
					Type:        productType,
					ReceptionId: receptionID,
				},
			)
		} else {
			newReception := struct {
				Reception md.Reception `json:"reception"`
				Products  []md.Product `json:"products"`
			}{
				Reception: md.Reception{
					ID:       receptionID,
					DateTime: receptionDate,
					PVZID:    pvzID,
					Status:   receptionStatus,
				},
				Products: []md.Product{
					{
						ID:          productID,
						DateTime:    productDate,
						Type:        productType,
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

	result := make([]*dto.GetPVZResponse, 0, len(pvzMap))
	for _, pvz := range pvzMap {
		result = append(result, pvz)
	}
	return result, nil
}

func (r *Repository) CloseLastReception(ctx context.Context, id uuid.UUID) (*md.Reception, error) {
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

	var reception md.Reception
	err = tx.GetContext(ctx, &reception, findLastReception, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repo.ErrReceptionAlreadyClosed
		}
		return nil, err
	}

	_, err = tx.ExecContext(ctx, closeReception, reception.ID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &reception, nil
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
	err = tx.GetContext(ctx, &reception, findLastReceptionForUpdate, id)
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

func (r *Repository) CreateReception(ctx context.Context, req *dto.CreateReceptionRequest) (*dto.CreateReceptionResponse, error) {
	tx, err := r.conn.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func(tx *sqlx.Tx) {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			zap.L().Error("Failed to rollback transaction", zap.Error(err))
		}
	}(tx)

	var res dto.CreateReceptionResponse
	err = tx.GetContext(ctx, &res, findLastReceptionForUpdate, req.PVZID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	} else {
		return nil, repo.ErrReceptionStillOpen
	}

	err = r.conn.GetContext(ctx, &res, createReception, req.PVZID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *Repository) AddItemToReception(ctx context.Context, req *dto.AddItemRequest) (*dto.AddItemResponse, error) {
	var reception md.Reception
	err := r.conn.GetContext(ctx, &reception, findLastReception, req.PVZID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repo.ErrNoActiveReception
		}
		return nil, err
	}

	var res dto.AddItemResponse
	err = r.conn.GetContext(ctx, &res, addItemToReception, reception.ID, req.Type)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "22P02" {
				return nil, repo.ErrTypeIsNotValid
			}
		}
		return nil, err
	}

	return &res, nil
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
