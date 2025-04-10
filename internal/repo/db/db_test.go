package db

import (
	"context"
	"database/sql"
	"errors"
	dto "github.com/JMURv/avito-spring/internal/dto/gen"
	md "github.com/JMURv/avito-spring/internal/models"
	repo2 "github.com/JMURv/avito-spring/internal/repo"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
	"regexp"
	"testing"
	"time"
)

func TestRepository_GetUserByEmail(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	repo := Repository{conn: db}
	ctx := context.Background()
	email := "test@example.com"

	testErr := errors.New("test-error")
	testUser := md.User{
		ID:       uuid.New(),
		Email:    email,
		Password: "hashedpassword",
		Role:     "admin",
	}

	tests := []struct {
		name       string
		setup      func()
		wantErr    error
		wantNilRes bool
	}{
		{
			name: "Success",
			setup: func() {
				rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "role"}).
					AddRow(testUser.ID.String(), testUser.Email, testUser.Password, testUser.Role)

				mock.ExpectQuery(regexp.QuoteMeta(getUserByEmail)).
					WithArgs(email).
					WillReturnRows(rows)
			},
			wantErr:    nil,
			wantNilRes: false,
		},
		{
			name: "NotFound",
			setup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(getUserByEmail)).
					WithArgs(email).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr:    repo2.ErrNotFound,
			wantNilRes: true,
		},
		{
			name: "DB Error",
			setup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(getUserByEmail)).
					WithArgs(email).
					WillReturnError(testErr)
			},
			wantErr:    testErr,
			wantNilRes: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.setup()
				res, err := repo.GetUserByEmail(ctx, email)

				if tt.wantErr != nil {
					require.Error(t, err)
					require.ErrorContains(t, err, tt.wantErr.Error())
				} else {
					require.NoError(t, err)
				}

				if tt.wantNilRes {
					require.Nil(t, res)
				} else {
					require.NotNil(t, res)
					require.Equal(t, testUser.Email, res.Email)
				}
			},
		)
	}
}

func TestRepository_CreateUser(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	repo := Repository{conn: db}
	ctx := context.Background()

	testID := uuid.New()
	testErr := errors.New("insert error")

	req := &dto.RegisterPostReq{
		Email:    "test@example.com",
		Password: "securehash",
		Role:     "admin",
	}

	tests := []struct {
		name    string
		setup   func()
		wantErr error
		wantID  uuid.UUID
	}{
		{
			name: "Success",
			setup: func() {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(testID.String())
				mock.ExpectQuery(regexp.QuoteMeta(createUser)).
					WithArgs(req.Email, req.Password, req.Role).
					WillReturnRows(rows)
			},
			wantErr: nil,
			wantID:  testID,
		},
		{
			name: "QueryError",
			setup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(createUser)).
					WithArgs(req.Email, req.Password, req.Role).
					WillReturnError(testErr)
			},
			wantErr: testErr,
			wantID:  uuid.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.setup()
				id, err := repo.CreateUser(ctx, req)

				if tt.wantErr != nil {
					require.Error(t, err)
					require.ErrorContains(t, err, tt.wantErr.Error())
					require.Equal(t, uuid.Nil, id)
				} else {
					require.NoError(t, err)
					require.Equal(t, tt.wantID, id)
				}
			},
		)
	}
}

func TestRepository_CreatePVZ(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	repo := Repository{conn: db}
	ctx := context.Background()

	testID := uuid.New()
	testTime := time.Now()
	testCity := "Moscow"
	testErr := errors.New("unexpected db error")

	tests := []struct {
		name     string
		setup    func()
		req      *dto.PVZ
		wantID   uuid.UUID
		wantTime time.Time
		wantErr  error
	}{
		{
			name: "Success",
			req:  &dto.PVZ{City: dto.PVZCity(testCity)},
			setup: func() {
				rows := sqlmock.NewRows([]string{"id", "created_at"}).
					AddRow(testID.String(), testTime)

				mock.ExpectQuery(regexp.QuoteMeta(createPVZ)).
					WithArgs(testCity).
					WillReturnRows(rows)
			},
			wantID:   testID,
			wantTime: testTime,
			wantErr:  nil,
		},
		{
			name: "InvalidCity_PG_22P02",
			req:  &dto.PVZ{City: "123_invalid"},
			setup: func() {
				pgErr := &pgconn.PgError{Code: "22P02"}
				mock.ExpectQuery(regexp.QuoteMeta(createPVZ)).
					WithArgs("123_invalid").
					WillReturnError(pgErr)
			},
			wantID:   uuid.Nil,
			wantTime: time.Time{},
			wantErr:  repo2.ErrCityIsNotValid,
		},
		{
			name: "Generic DB Error",
			req:  &dto.PVZ{City: "St.Petersburg"},
			setup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(createPVZ)).
					WithArgs("St.Petersburg").
					WillReturnError(testErr)
			},
			wantID:   uuid.Nil,
			wantTime: time.Time{},
			wantErr:  testErr,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.setup()
				id, createdAt, err := repo.CreatePVZ(ctx, tt.req)

				if tt.wantErr != nil {
					require.Error(t, err)
					require.ErrorIs(t, err, tt.wantErr)
				} else {
					require.NoError(t, err)
				}

				require.Equal(t, tt.wantID, id)
				require.Equal(t, tt.wantTime, createdAt)
			},
		)
	}
}

func TestRepository_GetPVZ(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	repo := Repository{conn: db}
	ctx := context.Background()

	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()
	page := int64(1)
	limit := int64(10)

	testPVZID := uuid.New().String()
	testReceptionID := uuid.New().String()
	testProductID := uuid.New().String()

	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name: "Success with data",
			setup: func() {
				rows := sqlmock.NewRows(
					[]string{
						"pickup_point_id", "pvz_city", "pvz_created_at",
						"reception_id", "reception_date", "reception_status",
						"product_id", "product_date", "product_type",
					},
				).AddRow(
					testPVZID, "Moscow", time.Now(),
					testReceptionID, time.Now(), "open",
					testProductID, time.Now(), "electronics",
				)

				mock.ExpectQuery(regexp.QuoteMeta(getPVZ)).
					WithArgs(start, end, limit, (page-1)*limit).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "DB error",
			setup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(getPVZ)).
					WithArgs(start, end, limit, (page-1)*limit).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "Scan error",
			setup: func() {
				rows := sqlmock.NewRows(
					[]string{
						"pickup_point_id", "pvz_city", "pvz_created_at",
						"reception_id", "reception_date", "reception_status",
						"product_id", "product_date", "product_type",
					},
				).AddRow(
					"invalid-uuid", "Moscow", time.Now(),
					testReceptionID, time.Now(), "open",
					testProductID, time.Now(), "electronics",
				)

				mock.ExpectQuery(regexp.QuoteMeta(getPVZ)).
					WithArgs(start, end, limit, (page-1)*limit).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.setup()
				res, err := repo.GetPVZ(ctx, page, limit, start, end)
				if tt.wantErr {
					require.Error(t, err)
					require.Nil(t, res)
				} else {
					require.NoError(t, err)
					require.NotNil(t, res)
					require.GreaterOrEqual(t, len(res), 1)
				}
			},
		)
	}
}

func TestRepository_CloseLastReception(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	repo := Repository{conn: db}
	ctx := context.Background()

	receptionID := uuid.New()
	testReception := dto.Reception{
		ID: dto.OptUUID{
			Value: receptionID,
			Set:   true,
		},
		DateTime: time.Now(),
		PvzId:    uuid.New(),
		Status:   "in_progress",
	}

	tests := []struct {
		name       string
		setup      func()
		wantErr    error
		wantNilRes bool
	}{
		{
			name: "Success",
			setup: func() {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id", "created_at", "pickup_point_id", "status"}).
					AddRow(
						testReception.ID.Value.String(),
						testReception.DateTime,
						testReception.PvzId.String(),
						testReception.Status,
					)

				mock.ExpectQuery(regexp.QuoteMeta(findLastReceptionForUpdate)).
					WithArgs(receptionID).
					WillReturnRows(rows)

				mock.ExpectExec(regexp.QuoteMeta(closeReception)).
					WithArgs(testReception.ID.Value.String()).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit()
			},
			wantErr:    nil,
			wantNilRes: false,
		},
		{
			name: "ReceptionAlreadyClosed (no rows)",
			setup: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(findLastReceptionForUpdate)).
					WithArgs(receptionID).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectRollback()
			},
			wantErr:    repo2.ErrReceptionAlreadyClosed,
			wantNilRes: true,
		},
		{
			name: "GetContext error",
			setup: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(findLastReceptionForUpdate)).
					WithArgs(receptionID).
					WillReturnError(errors.New("db get error"))
				mock.ExpectRollback()
			},
			wantErr:    errors.New("db get error"),
			wantNilRes: true,
		},
		{
			name: "ExecContext error",
			setup: func() {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id", "created_at", "pickup_point_id", "status"}).
					AddRow(
						testReception.ID.Value.String(),
						testReception.DateTime,
						testReception.PvzId.String(),
						testReception.Status,
					)

				mock.ExpectQuery(regexp.QuoteMeta(findLastReceptionForUpdate)).
					WithArgs(receptionID).
					WillReturnRows(rows)

				mock.ExpectExec(regexp.QuoteMeta(closeReception)).
					WithArgs(testReception.ID.Value.String()).
					WillReturnError(errors.New("exec error"))
				mock.ExpectRollback()
			},
			wantErr:    errors.New("exec error"),
			wantNilRes: true,
		},
		{
			name: "Commit error",
			setup: func() {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id", "created_at", "pickup_point_id", "status"}).
					AddRow(
						testReception.ID.Value.String(),
						testReception.DateTime,
						testReception.PvzId.String(),
						testReception.Status,
					)

				mock.ExpectQuery(regexp.QuoteMeta(findLastReceptionForUpdate)).
					WithArgs(receptionID).
					WillReturnRows(rows)

				mock.ExpectExec(regexp.QuoteMeta(closeReception)).
					WithArgs(testReception.ID.Value.String()).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit().WillReturnError(errors.New("commit error"))
			},
			wantErr:    errors.New("commit error"),
			wantNilRes: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.setup()
				res, err := repo.CloseLastReception(ctx, receptionID)

				if tt.wantErr != nil {
					require.Error(t, err)
					require.ErrorContains(t, err, tt.wantErr.Error())
				} else {
					require.NoError(t, err)
				}

				if tt.wantNilRes {
					require.Nil(t, res)
				} else {
					require.NotNil(t, res)
					require.Equal(t, testReception.ID, res.ID)
				}

				require.NoError(t, mock.ExpectationsWereMet())
			},
		)
	}
}

func TestRepository_DeleteLastProduct(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	repo := Repository{conn: db}
	ctx := context.Background()

	receptionID := uuid.New()
	testReception := md.Reception{
		ID:       receptionID,
		DateTime: time.Now(),
		PVZID:    uuid.New(),
		Status:   "open",
	}

	tests := []struct {
		name    string
		setup   func()
		wantErr error
	}{
		{
			name: "Success",
			setup: func() {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id", "created_at", "pickup_point_id", "status"}).
					AddRow(
						testReception.ID.String(),
						testReception.DateTime,
						testReception.PVZID.String(),
						testReception.Status,
					)
				mock.ExpectQuery(regexp.QuoteMeta(findLastReception)).
					WithArgs(receptionID).
					WillReturnRows(rows)

				mock.ExpectExec(regexp.QuoteMeta(deleteLastProduct)).
					WithArgs(testReception.ID).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit()
			},
			wantErr: nil,
		},
		{
			name: "NoActiveReception",
			setup: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(findLastReception)).
					WithArgs(receptionID).
					WillReturnError(sql.ErrNoRows)
				mock.ExpectRollback()
			},
			wantErr: repo2.ErrNoActiveReception,
		},
		{
			name: "NoItemsAffected",
			setup: func() {
				mock.ExpectBegin()

				rows := sqlmock.NewRows([]string{"id", "created_at", "pickup_point_id", "status"}).
					AddRow(
						testReception.ID.String(),
						testReception.DateTime,
						testReception.PVZID.String(),
						testReception.Status,
					)
				mock.ExpectQuery(regexp.QuoteMeta(findLastReception)).
					WithArgs(receptionID).
					WillReturnRows(rows)

				mock.ExpectExec(regexp.QuoteMeta(deleteLastProduct)).
					WithArgs(testReception.ID).
					WillReturnResult(sqlmock.NewResult(0, 0))

				mock.ExpectRollback()
			},
			wantErr: repo2.ErrNoItems,
		},
		{
			name: "GetContextError",
			setup: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(regexp.QuoteMeta(findLastReception)).
					WithArgs(receptionID).
					WillReturnError(errors.New("db get error"))
				mock.ExpectRollback()
			},
			wantErr: errors.New("db get error"),
		},
		{
			name: "ExecContextError",
			setup: func() {
				mock.ExpectBegin()

				rows := sqlmock.NewRows([]string{"id", "created_at", "pickup_point_id", "status"}).
					AddRow(
						testReception.ID.String(),
						testReception.DateTime,
						testReception.PVZID.String(),
						testReception.Status,
					)
				mock.ExpectQuery(regexp.QuoteMeta(findLastReception)).
					WithArgs(receptionID).
					WillReturnRows(rows)

				mock.ExpectExec(regexp.QuoteMeta(deleteLastProduct)).
					WithArgs(testReception.ID).
					WillReturnError(errors.New("exec error"))

				mock.ExpectRollback()
			},
			wantErr: errors.New("exec error"),
		},
		{
			name: "CommitError",
			setup: func() {
				mock.ExpectBegin()
				rows := sqlmock.NewRows([]string{"id", "created_at", "pickup_point_id", "status"}).
					AddRow(
						testReception.ID.String(),
						testReception.DateTime,
						testReception.PVZID.String(),
						testReception.Status,
					)
				mock.ExpectQuery(regexp.QuoteMeta(findLastReception)).
					WithArgs(receptionID).
					WillReturnRows(rows)

				mock.ExpectExec(regexp.QuoteMeta(deleteLastProduct)).
					WithArgs(testReception.ID).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit().WillReturnError(errors.New("commit error"))
			},
			wantErr: errors.New("commit error"),
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.setup()
				err := repo.DeleteLastProduct(ctx, receptionID)

				if tt.wantErr != nil {
					require.Error(t, err)
					require.ErrorContains(t, err, tt.wantErr.Error())
				} else {
					require.NoError(t, err)
				}

				require.NoError(t, mock.ExpectationsWereMet())
			},
		)
	}
}

func TestRepository_CreateReception(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	repo := Repository{conn: db}
	ctx := context.Background()

	req := &dto.ReceptionsPostReq{
		PvzId: uuid.New(),
	}
	testResponse := dto.Reception{
		ID: dto.OptUUID{
			Value: uuid.New(),
			Set:   true,
		},
		PvzId:    req.PvzId,
		Status:   "open",
		DateTime: time.Now(),
	}

	tests := []struct {
		name       string
		setup      func()
		wantErr    error
		wantResult *dto.Reception
	}{
		{
			name: "Success",
			setup: func() {
				mock.ExpectBegin()

				mock.ExpectQuery(regexp.QuoteMeta(findLastReceptionForUpdate)).
					WithArgs(req.PvzId).
					WillReturnError(sql.ErrNoRows)

				mock.ExpectQuery(regexp.QuoteMeta(createReception)).
					WithArgs(req.PvzId).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "pickup_point_id", "status", "created_at"}).
							AddRow(
								testResponse.ID.Value.String(),
								req.PvzId.String(),
								testResponse.Status,
								testResponse.DateTime,
							),
					)

				mock.ExpectCommit()
			},
			wantErr:    nil,
			wantResult: &testResponse,
		},
		{
			name: "ReceptionAlreadyOpen",
			setup: func() {
				mock.ExpectBegin()

				mock.ExpectQuery(regexp.QuoteMeta(findLastReceptionForUpdate)).
					WithArgs(req.PvzId).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "pickup_point_id", "status", "created_at"}).
							AddRow(
								testResponse.ID.Value.String(),
								req.PvzId.String(),
								"in_progress",
								testResponse.DateTime,
							),
					)

				mock.ExpectRollback()
			},
			wantErr:    repo2.ErrReceptionStillOpen,
			wantResult: nil,
		},
		{
			name: "FindLastReceptionError",
			setup: func() {
				mock.ExpectBegin()

				mock.ExpectQuery(regexp.QuoteMeta(findLastReceptionForUpdate)).
					WithArgs(req.PvzId).
					WillReturnError(errors.New("db query error"))

				mock.ExpectRollback()
			},
			wantErr:    errors.New("db query error"),
			wantResult: nil,
		},
		{
			name: "CreateReceptionError",
			setup: func() {
				mock.ExpectBegin()

				mock.ExpectQuery(regexp.QuoteMeta(findLastReceptionForUpdate)).
					WithArgs(req.PvzId).
					WillReturnError(sql.ErrNoRows)

				mock.ExpectQuery(regexp.QuoteMeta(createReception)).
					WithArgs(req.PvzId).
					WillReturnError(errors.New("db create error"))

				mock.ExpectRollback()
			},
			wantErr:    errors.New("db create error"),
			wantResult: nil,
		},
		{
			name: "CommitError",
			setup: func() {
				mock.ExpectBegin()

				mock.ExpectQuery(regexp.QuoteMeta(findLastReceptionForUpdate)).
					WithArgs(req.PvzId).
					WillReturnError(sql.ErrNoRows)

				mock.ExpectQuery(regexp.QuoteMeta(createReception)).
					WithArgs(req.PvzId).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "pickup_point_id", "status", "created_at"}).
							AddRow(
								testResponse.ID.Value.String(),
								req.PvzId.String(),
								testResponse.Status,
								testResponse.DateTime,
							),
					)

				mock.ExpectCommit().WillReturnError(errors.New("commit error"))
			},
			wantErr:    errors.New("commit error"),
			wantResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.setup()
				res, err := repo.CreateReception(ctx, req)

				if tt.wantErr != nil {
					require.Error(t, err)
					require.ErrorContains(t, err, tt.wantErr.Error())
				} else {
					require.NoError(t, err)
				}

				if tt.wantResult != nil {
					require.Equal(t, tt.wantResult, res)
				} else {
					require.Nil(t, res)
				}

				require.NoError(t, mock.ExpectationsWereMet())
			},
		)
	}
}

func TestRepository_AddItemToReception(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	repo := Repository{conn: db}
	ctx := context.Background()

	pvzID := uuid.New()
	receptionID := uuid.New()

	req := &dto.ProductsPostReq{
		PvzId: pvzID,
		Type:  "fragile",
	}
	testReception := md.Reception{
		ID:       receptionID,
		DateTime: time.Now(),
		PVZID:    pvzID,
		Status:   "open",
	}

	testResp := dto.Product{
		ID: dto.OptUUID{
			Value: uuid.New(),
			Set:   true,
		},
		Type:        dto.ProductType(req.Type),
		ReceptionId: receptionID,
		DateTime: dto.OptDateTime{
			Value: time.Now(),
			Set:   true,
		},
	}

	tests := []struct {
		name       string
		setup      func()
		wantErr    error
		wantResult *dto.Product
	}{
		{
			name: "Success",
			setup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(findLastReception)).
					WithArgs(pvzID).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "created_at", "pickup_point_id", "status"}).
							AddRow(
								testReception.ID.String(),
								testReception.DateTime,
								testReception.PVZID.String(),
								testReception.Status,
							),
					)

				mock.ExpectQuery(regexp.QuoteMeta(addItemToReception)).
					WithArgs(receptionID, req.Type).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "type", "reception_id", "created_at"}).
							AddRow(
								testResp.ID.Value.String(),
								testResp.Type,
								testResp.ReceptionId.String(),
								testResp.DateTime.Value,
							),
					)
			},
			wantErr:    nil,
			wantResult: &testResp,
		},
		{
			name: "No Active Reception",
			setup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(findLastReception)).
					WithArgs(pvzID).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr:    repo2.ErrNoActiveReception,
			wantResult: nil,
		},
		{
			name: "Find Reception DB Error",
			setup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(findLastReception)).
					WithArgs(pvzID).
					WillReturnError(errors.New("db error"))
			},
			wantErr:    errors.New("db error"),
			wantResult: nil,
		},
		{
			name: "Invalid Type Error",
			setup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(findLastReception)).
					WithArgs(pvzID).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "created_at", "pickup_point_id", "status"}).
							AddRow(
								testReception.ID.String(),
								testReception.DateTime,
								testReception.PVZID.String(),
								testReception.Status,
							),
					)

				mock.ExpectQuery(regexp.QuoteMeta(addItemToReception)).
					WithArgs(receptionID, req.Type).
					WillReturnError(&pgconn.PgError{Code: "22P02"})
			},
			wantErr:    repo2.ErrTypeIsNotValid,
			wantResult: nil,
		},
		{
			name: "Insert Item DB Error",
			setup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(findLastReception)).
					WithArgs(pvzID).
					WillReturnRows(
						sqlmock.NewRows([]string{"id", "created_at", "pickup_point_id", "status"}).
							AddRow(
								testReception.ID.String(),
								testReception.DateTime,
								testReception.PVZID.String(),
								testReception.Status,
							),
					)

				mock.ExpectQuery(regexp.QuoteMeta(addItemToReception)).
					WithArgs(receptionID, req.Type).
					WillReturnError(errors.New("insert error"))
			},
			wantErr:    errors.New("insert error"),
			wantResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.setup()
				res, err := repo.AddItemToReception(ctx, req)

				if tt.wantErr != nil {
					require.Error(t, err)
					require.ErrorContains(t, err, tt.wantErr.Error())
				} else {
					require.NoError(t, err)
				}

				if tt.wantResult != nil {
					require.Equal(t, tt.wantResult, res)
				} else {
					require.Nil(t, res)
				}

				require.NoError(t, mock.ExpectationsWereMet())
			},
		)
	}
}

func TestRepository_GetPVZList(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	repo := Repository{conn: db}
	ctx := context.Background()

	testPVZ := md.PVZ{
		ID:               uuid.New(),
		City:             "Moscow",
		RegistrationDate: time.Now(),
	}

	tests := []struct {
		name       string
		setup      func()
		wantErr    error
		wantResult []*md.PVZ
	}{
		{
			name: "Success",
			setup: func() {
				rows := sqlmock.NewRows([]string{"id", "city", "created_at"}).
					AddRow(testPVZ.ID.String(), testPVZ.City, testPVZ.RegistrationDate)

				mock.ExpectQuery(regexp.QuoteMeta(listPVZs)).
					WillReturnRows(rows)
			},
			wantErr:    nil,
			wantResult: []*md.PVZ{&testPVZ},
		},
		{
			name: "No Rows",
			setup: func() {
				rows := sqlmock.NewRows([]string{"id", "city", "created_at"})
				mock.ExpectQuery(regexp.QuoteMeta(listPVZs)).
					WillReturnRows(rows)
			},
			wantErr:    nil,
			wantResult: nil,
		},
		{
			name: "DB Error",
			setup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(listPVZs)).
					WillReturnError(errors.New("db error"))
			},
			wantErr:    errors.New("db error"),
			wantResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.setup()
				res, err := repo.GetPVZList(ctx)

				if tt.wantErr != nil {
					require.Error(t, err)
					require.ErrorContains(t, err, tt.wantErr.Error())
				} else {
					require.NoError(t, err)
					require.Equal(t, tt.wantResult, res)
				}

				require.NoError(t, mock.ExpectationsWereMet())
			},
		)
	}
}
