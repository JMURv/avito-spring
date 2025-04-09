package http

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/JMURv/avito-spring/internal/auth"
	"github.com/JMURv/avito-spring/internal/config"
	"github.com/JMURv/avito-spring/internal/ctrl"
	"github.com/JMURv/avito-spring/internal/dto"
	hdl "github.com/JMURv/avito-spring/internal/hdl/http"
	"github.com/JMURv/avito-spring/internal/repo/db"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const configPath = "../../../configs/test.config.yaml"
const getTables = `
SELECT tablename 
FROM pg_tables 
WHERE schemaname = 'public';
`

func setupTestServer() (*httptest.Server, func()) {
	zap.ReplaceGlobals(zap.Must(zap.NewDevelopment()))

	conf := config.MustLoad(configPath)
	au := auth.New(conf)
	repo := db.New(conf)
	svc := ctrl.New(repo, au)
	h := hdl.New(svc, au)

	mux := http.NewServeMux()
	hdl.RegisterRoutes(mux, h, au)

	cleanupFunc := func() {
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

		rows, err := conn.Query(getTables)
		if err != nil {
			zap.L().Fatal("Failed to fetch table names", zap.Error(err))
		}
		defer func(rows *sql.Rows) {
			if err := rows.Close(); err != nil {
				zap.L().Debug("Error while closing rows", zap.Error(err))
			}
		}(rows)

		var tables []string
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				zap.L().Fatal("Failed to scan table name", zap.Error(err))
			}
			tables = append(tables, name)
		}

		if len(tables) == 0 {
			return
		}

		_, err = conn.Exec(fmt.Sprintf("TRUNCATE TABLE %v RESTART IDENTITY CASCADE;", strings.Join(tables, ", ")))
		if err != nil {
			zap.L().Fatal("Failed to truncate tables", zap.Error(err))
		}
	}

	return httptest.NewServer(mux), cleanupFunc
}

func TestFullReceptionFlow(t *testing.T) {
	srv, cleanup := setupTestServer()
	t.Cleanup(cleanup)

	client := srv.Client()

	// Регистрация модератора
	registerMod := dto.RegisterRequest{
		Email:    "mod@avito.ru",
		Password: "password",
		Role:     "moderator",
	}
	buf, err := json.Marshal(registerMod)
	require.NoError(t, err)

	resp, err := client.Post(srv.URL+"/register", "application/json", bytes.NewReader(buf))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	// Логин модератора
	loginMod := dto.LoginRequest{
		Email:    "mod@avito.ru",
		Password: "password",
	}
	buf, err = json.Marshal(loginMod)
	require.NoError(t, err)

	resp, err = client.Post(srv.URL+"/login", "application/json", bytes.NewReader(buf))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var loginResp dto.LoginResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&loginResp))
	resp.Body.Close()

	authHeader := "Bearer " + loginResp.Token

	// Создание ПВЗ
	pvzReq := dto.CreatePVZRequest{City: "Москва"}
	buf, err = json.Marshal(pvzReq)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, srv.URL+"/pvz", bytes.NewReader(buf))
	require.NoError(t, err)

	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var pvzRes dto.CreatePVZResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pvzRes))
	resp.Body.Close()

	// Регистрация сотрудника
	registerEmp := dto.RegisterRequest{
		Email:    "emp@avito.ru",
		Password: "password",
		Role:     "employee",
	}
	buf, err = json.Marshal(registerEmp)
	require.NoError(t, err)
	resp, err = client.Post(srv.URL+"/register", "application/json", bytes.NewReader(buf))
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	// Логин сотрудника
	loginEmp := dto.LoginRequest{
		Email:    "emp@avito.ru",
		Password: "password",
	}
	buf, err = json.Marshal(loginEmp)
	require.NoError(t, err)
	resp, err = client.Post(srv.URL+"/login", "application/json", bytes.NewReader(buf))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&loginResp))
	resp.Body.Close()
	authHeader = "Bearer " + loginResp.Token

	// 6. Создание приёмки
	recReq := dto.CreateReceptionRequest{PVZID: pvzRes.ID}
	buf, err = json.Marshal(recReq)
	require.NoError(t, err)
	req, err = http.NewRequest(http.MethodPost, srv.URL+"/receptions", bytes.NewReader(buf))
	require.NoError(t, err)
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// Добавление 50 товаров
	for i := 0; i < 50; i++ {
		addReq := dto.AddItemRequest{
			Type:  "электроника",
			PVZID: pvzRes.ID,
		}
		buf, err = json.Marshal(addReq)
		require.NoError(t, err)
		req, err = http.NewRequest(http.MethodPost, srv.URL+"/products", bytes.NewReader(buf))
		require.NoError(t, err)
		req.Header.Set("Authorization", authHeader)
		req.Header.Set("Content-Type", "application/json")
		resp, err = client.Do(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		resp.Body.Close()
	}

	// Закрытие приёмки
	url := fmt.Sprintf("/pvz/%s/close_last_reception", pvzRes.ID.String())
	req, err = http.NewRequest(http.MethodPost, srv.URL+url, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", authHeader)
	resp, err = client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}
