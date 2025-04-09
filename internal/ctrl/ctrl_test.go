package ctrl

import (
	"context"
	"errors"
	"github.com/JMURv/avito-spring/internal/auth"
	"github.com/JMURv/avito-spring/internal/dto"
	md "github.com/JMURv/avito-spring/internal/models"
	"github.com/JMURv/avito-spring/internal/repo"
	"github.com/JMURv/avito-spring/tests/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestController_DummyLogin(t *testing.T) {
	ctx := context.Background()
	mock := gomock.NewController(t)
	defer mock.Finish()

	auth := mocks.NewMockCore(mock)
	repo := mocks.NewMockAppRepo(mock)
	ctrl := New(repo, auth)
	testErr := errors.New("test-err")

	tests := []struct {
		name       string
		req        *dto.DummyLoginRequest
		expect     func()
		assertions func(res *dto.DummyLoginResponse, err error)
	}{
		{
			name: "NewToken Err",
			req: &dto.DummyLoginRequest{
				Role: "role",
			},
			assertions: func(res *dto.DummyLoginResponse, err error) {
				assert.Nil(t, res)
				assert.Equal(t, testErr, err)
			},
			expect: func() {
				auth.EXPECT().NewToken(gomock.Any(), gomock.Any()).Return("", testErr)
			},
		},
		{
			name: "Success",
			req: &dto.DummyLoginRequest{
				Role: "role",
			},
			assertions: func(res *dto.DummyLoginResponse, err error) {
				assert.Nil(t, err)
				assert.Equal(t, "token", res.Token)
			},
			expect: func() {
				auth.EXPECT().NewToken(gomock.Any(), gomock.Any()).Return("token", nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.expect()
				res, err := ctrl.DummyLogin(ctx, tt.req)
				tt.assertions(res, err)
			},
		)
	}
}

func TestController_Login(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	authMock := mocks.NewMockCore(mockCtrl)
	repoMock := mocks.NewMockAppRepo(mockCtrl)
	ctrl := New(repoMock, authMock)

	testErr := errors.New("test error")

	tests := []struct {
		name       string
		req        *dto.LoginRequest
		expect     func()
		assertions func(*dto.LoginResponse, error)
	}{
		{
			name: "User not found",
			req: &dto.LoginRequest{
				Email:    "notfound@example.com",
				Password: "password",
			},
			expect: func() {
				repoMock.EXPECT().GetUserByEmail(ctx, "notfound@example.com").Return(nil, repo.ErrNotFound)
			},
			assertions: func(res *dto.LoginResponse, err error) {
				assert.Nil(t, res)
				assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
			},
		},
		{
			name: "GetUserByEmail error",
			req: &dto.LoginRequest{
				Email:    "error@example.com",
				Password: "password",
			},
			expect: func() {
				repoMock.EXPECT().GetUserByEmail(ctx, "error@example.com").Return(nil, testErr)
			},
			assertions: func(res *dto.LoginResponse, err error) {
				assert.Nil(t, res)
				assert.Equal(t, testErr, err)
			},
		},
		{
			name: "Invalid password",
			req: &dto.LoginRequest{
				Email:    "user@example.com",
				Password: "wrongpass",
			},
			expect: func() {
				repoMock.EXPECT().GetUserByEmail(ctx, "user@example.com").Return(
					&md.User{
						Password: "hashedpass",
					}, nil,
				)
				authMock.EXPECT().ComparePasswords(
					[]byte("hashedpass"),
					[]byte("wrongpass"),
				).Return(auth.ErrInvalidCredentials)
			},
			assertions: func(res *dto.LoginResponse, err error) {
				assert.Nil(t, res)
				assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
			},
		},
		{
			name: "ComparePasswords error",
			req: &dto.LoginRequest{
				Email:    "user@example.com",
				Password: "pass",
			},
			expect: func() {
				repoMock.EXPECT().GetUserByEmail(ctx, "user@example.com").Return(
					&md.User{
						Password: "hashed",
					}, nil,
				)
				authMock.EXPECT().ComparePasswords([]byte("hashed"), []byte("pass")).Return(testErr)
			},
			assertions: func(res *dto.LoginResponse, err error) {
				assert.Nil(t, res)
				assert.Equal(t, testErr, err)
			},
		},
		{
			name: "NewToken error",
			req: &dto.LoginRequest{
				Email:    "user@example.com",
				Password: "correctpass",
			},
			expect: func() {
				repoMock.EXPECT().GetUserByEmail(ctx, "user@example.com").Return(
					&md.User{
						ID:       uuid.New(),
						Password: "hashedpass",
						Role:     "user",
					}, nil,
				)
				authMock.EXPECT().ComparePasswords([]byte("hashedpass"), []byte("correctpass")).Return(nil)
				authMock.EXPECT().NewToken(gomock.Any(), "user").Return("", testErr)
			},
			assertions: func(res *dto.LoginResponse, err error) {
				assert.Nil(t, res)
				assert.Equal(t, testErr, err)
			},
		},
		{
			name: "Success",
			req: &dto.LoginRequest{
				Email:    "success@example.com",
				Password: "correctpass",
			},
			expect: func() {
				repoMock.EXPECT().GetUserByEmail(ctx, "success@example.com").Return(
					&md.User{
						ID:       uuid.New(),
						Password: "hashedcorrect",
						Role:     "admin",
					}, nil,
				)
				authMock.EXPECT().ComparePasswords([]byte("hashedcorrect"), []byte("correctpass")).Return(nil)
				authMock.EXPECT().NewToken(gomock.Any(), "admin").Return("valid-token", nil)
			},
			assertions: func(res *dto.LoginResponse, err error) {
				assert.NoError(t, err)
				assert.Equal(t, "valid-token", res.Token)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if tt.expect != nil {
					tt.expect()
				}
				res, err := ctrl.Login(ctx, tt.req)
				tt.assertions(res, err)
			},
		)
	}
}

func TestController_Register(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	authMock := mocks.NewMockCore(mockCtrl)
	repoMock := mocks.NewMockAppRepo(mockCtrl)
	ctrl := New(repoMock, authMock)

	testErr := errors.New("test error")
	testID := uuid.New()

	tests := []struct {
		name       string
		req        *dto.RegisterRequest
		expect     func()
		assertions func(*dto.RegisterResponse, error)
	}{
		{
			name: "Hashing error",
			req: &dto.RegisterRequest{
				Password: "password",
			},
			expect: func() {
				authMock.EXPECT().Hash("password").Return("", testErr)
			},
			assertions: func(res *dto.RegisterResponse, err error) {
				assert.Nil(t, res)
				assert.ErrorIs(t, err, testErr)
			},
		},
		{
			name: "CreateUser error",
			req: &dto.RegisterRequest{
				Email:    "error@example.com",
				Password: "password",
				Role:     "user",
			},
			expect: func() {
				authMock.EXPECT().Hash("password").Return("hashedpass", nil)
				repoMock.EXPECT().CreateUser(
					ctx,
					&dto.RegisterRequest{
						Email:    "error@example.com",
						Password: "hashedpass",
						Role:     "user",
					},
				).Return(uuid.Nil, testErr)
			},
			assertions: func(res *dto.RegisterResponse, err error) {
				assert.Nil(t, res)
				assert.ErrorIs(t, err, testErr)
			},
		},
		{
			name: "Success",
			req: &dto.RegisterRequest{
				Email:    "success@example.com",
				Password: "password",
				Role:     "admin",
			},
			expect: func() {
				authMock.EXPECT().Hash("password").Return("hashedpass", nil)
				repoMock.EXPECT().CreateUser(
					ctx,
					&dto.RegisterRequest{
						Email:    "success@example.com",
						Password: "hashedpass",
						Role:     "admin",
					},
				).Return(testID, nil)
			},
			assertions: func(res *dto.RegisterResponse, err error) {
				assert.NoError(t, err)
				assert.Equal(t, testID, res.ID)
				assert.Equal(t, "success@example.com", res.Email)
				assert.Equal(t, "admin", res.Role)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if tt.expect != nil {
					tt.expect()
				}
				res, err := ctrl.Register(ctx, tt.req)
				tt.assertions(res, err)
			},
		)
	}
}

func TestController_GetPVZ(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	repoMock := mocks.NewMockAppRepo(mockCtrl)
	authMock := mocks.NewMockCore(mockCtrl)
	ctrl := New(repoMock, authMock)

	testErr := errors.New("test error")
	startDate, endDate := time.Now(), time.Now().Add(24*time.Hour)
	page, limit := int64(1), int64(10)

	sampleResponse := []*dto.GetPVZResponse{
		{
			PVZ: md.PVZ{},
			Receptions: []struct {
				Reception md.Reception `json:"reception"`
				Products  []md.Product `json:"products"`
			}{},
		},
	}

	tests := []struct {
		name       string
		expect     func()
		assertions func(res []*dto.GetPVZResponse, err error)
	}{
		{
			name: "GetPVZ returns error",
			expect: func() {
				repoMock.EXPECT().
					GetPVZ(ctx, page, limit, startDate, endDate).
					Return(nil, testErr)
			},
			assertions: func(res []*dto.GetPVZResponse, err error) {
				assert.Nil(t, res)
				assert.ErrorIs(t, err, testErr)
			},
		},
		{
			name: "Successful GetPVZ",
			expect: func() {
				repoMock.EXPECT().
					GetPVZ(ctx, page, limit, startDate, endDate).
					Return(sampleResponse, nil)
			},
			assertions: func(res []*dto.GetPVZResponse, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, res)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.expect()
				res, err := ctrl.GetPVZ(ctx, page, limit, startDate, endDate)
				tt.assertions(res, err)
			},
		)
	}
}

func TestController_CreatePVZ(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	repoMock := mocks.NewMockAppRepo(mockCtrl)
	authMock := mocks.NewMockCore(mockCtrl)
	ctrl := New(repoMock, authMock)

	testErr := errors.New("test error")
	invalidCityErr := repo.ErrCityIsNotValid

	testID := uuid.New()
	createdAt := time.Now()

	tests := []struct {
		name       string
		req        *dto.CreatePVZRequest
		expect     func()
		assertions func(resp *dto.CreatePVZResponse, err error)
	}{
		{
			name: "City not valid error",
			req: &dto.CreatePVZRequest{
				City: "InvalidCity",
			},
			expect: func() {
				repoMock.EXPECT().
					CreatePVZ(ctx, gomock.Any()).
					Return(uuid.Nil, time.Time{}, invalidCityErr)
			},
			assertions: func(resp *dto.CreatePVZResponse, err error) {
				assert.Nil(t, resp)
				assert.ErrorIs(t, err, ErrCityIsNotValid)
			},
		},
		{
			name: "General error",
			req: &dto.CreatePVZRequest{
				City: "AnyCity",
			},
			expect: func() {
				repoMock.EXPECT().
					CreatePVZ(ctx, gomock.Any()).
					Return(uuid.Nil, time.Time{}, testErr)
			},
			assertions: func(resp *dto.CreatePVZResponse, err error) {
				assert.Nil(t, resp)
				assert.Error(t, err)
			},
		},
		{
			name: "Successful creation",
			req: &dto.CreatePVZRequest{
				City: "TestCity",
			},
			expect: func() {
				repoMock.EXPECT().
					CreatePVZ(ctx, &dto.CreatePVZRequest{City: "TestCity"}).
					Return(testID, createdAt, nil)
			},
			assertions: func(resp *dto.CreatePVZResponse, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, testID, resp.ID)
				assert.Equal(t, createdAt, resp.RegistrationDate)
				assert.Equal(t, "TestCity", resp.City)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.expect()
				resp, err := ctrl.CreatePVZ(ctx, tt.req)
				tt.assertions(resp, err)
			},
		)
	}
}

func TestController_CloseLastReception(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	repoMock := mocks.NewMockAppRepo(mockCtrl)
	authMock := mocks.NewMockCore(mockCtrl)
	ctrl := New(repoMock, authMock)

	testErr := errors.New("test error")
	closedAlreadyErr := repo.ErrReceptionAlreadyClosed
	testID := uuid.New()
	now := time.Now()
	sampleReception := &md.Reception{
		ID:       testID,
		DateTime: now,
		PVZID:    testID,
		Status:   "closed",
	}

	tests := []struct {
		name       string
		id         uuid.UUID
		expect     func()
		assertions func(*md.Reception, error)
	}{
		{
			name: "Reception already closed",
			id:   testID,
			expect: func() {
				repoMock.EXPECT().
					CloseLastReception(ctx, testID).
					Return(nil, closedAlreadyErr)
			},
			assertions: func(res *md.Reception, err error) {
				assert.Nil(t, res)
				assert.ErrorIs(t, err, ErrReceptionAlreadyClosed)
			},
		},
		{
			name: "General error when closing reception",
			id:   testID,
			expect: func() {
				repoMock.EXPECT().
					CloseLastReception(ctx, testID).
					Return(nil, testErr)
			},
			assertions: func(res *md.Reception, err error) {
				assert.Nil(t, res)
				assert.Error(t, err)
			},
		},
		{
			name: "Successful close of last reception",
			id:   testID,
			expect: func() {
				repoMock.EXPECT().
					CloseLastReception(ctx, testID).
					Return(sampleReception, nil)
			},
			assertions: func(res *md.Reception, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, testID, res.ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.expect()
				res, err := ctrl.CloseLastReception(ctx, tt.id)
				tt.assertions(res, err)
			},
		)
	}
}

func TestController_DeleteLastProduct(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	repoMock := mocks.NewMockAppRepo(mockCtrl)
	authMock := mocks.NewMockCore(mockCtrl)
	ctrl := New(repoMock, authMock)

	testID := uuid.New()
	testErr := errors.New("test error")

	tests := []struct {
		name       string
		id         uuid.UUID
		expect     func()
		assertions func(err error)
	}{
		{
			name: "No active reception",
			id:   testID,
			expect: func() {
				repoMock.EXPECT().
					DeleteLastProduct(ctx, testID).
					Return(repo.ErrNoActiveReception)
			},
			assertions: func(err error) {
				assert.ErrorIs(t, err, ErrNoActiveReception)
			},
		},
		{
			name: "No items for deletion",
			id:   testID,
			expect: func() {
				repoMock.EXPECT().
					DeleteLastProduct(ctx, testID).
					Return(repo.ErrNoItems)
			},
			assertions: func(err error) {
				assert.ErrorIs(t, err, ErrNoItems)
			},
		},
		{
			name: "General error",
			id:   testID,
			expect: func() {
				repoMock.EXPECT().
					DeleteLastProduct(ctx, testID).
					Return(testErr)
			},
			assertions: func(err error) {
				assert.Error(t, err)
				assert.Equal(t, testErr, err)
			},
		},
		{
			name: "Successful deletion",
			id:   testID,
			expect: func() {
				repoMock.EXPECT().
					DeleteLastProduct(ctx, testID).
					Return(nil)
			},
			assertions: func(err error) {
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.expect()
				err := ctrl.DeleteLastProduct(ctx, tt.id)
				tt.assertions(err)
			},
		)
	}
}

func TestController_CreateReception(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	repoMock := mocks.NewMockAppRepo(mockCtrl)
	authMock := mocks.NewMockCore(mockCtrl)
	ctrl := New(repoMock, authMock)

	generalErr := errors.New("general error")
	testPVZID := uuid.New()
	now := time.Now()
	sampleResponse := &dto.CreateReceptionResponse{
		ID:       testPVZID,
		DateTime: now,
		PVZID:    testPVZID,
		Status:   "open",
	}

	tests := []struct {
		name       string
		req        *dto.CreateReceptionRequest
		expect     func()
		assertions func(resp *dto.CreateReceptionResponse, err error)
	}{
		{
			name: "Reception still open error",
			req: &dto.CreateReceptionRequest{
				PVZID: testPVZID,
			},
			expect: func() {
				repoMock.EXPECT().
					CreateReception(ctx, gomock.Any()).
					Return(nil, repo.ErrReceptionStillOpen)
			},
			assertions: func(resp *dto.CreateReceptionResponse, err error) {
				assert.Nil(t, resp)
				assert.ErrorIs(t, err, ErrReceptionStillOpen)
			},
		},
		{
			name: "General error",
			req: &dto.CreateReceptionRequest{
				PVZID: testPVZID,
			},
			expect: func() {
				repoMock.EXPECT().
					CreateReception(ctx, gomock.Any()).
					Return(nil, generalErr)
			},
			assertions: func(resp *dto.CreateReceptionResponse, err error) {
				assert.Nil(t, resp)
				assert.Equal(t, generalErr, err)
			},
		},
		{
			name: "Successful creation",
			req: &dto.CreateReceptionRequest{
				PVZID: testPVZID,
			},
			expect: func() {
				repoMock.EXPECT().
					CreateReception(ctx, &dto.CreateReceptionRequest{PVZID: testPVZID}).
					Return(sampleResponse, nil)
			},
			assertions: func(resp *dto.CreateReceptionResponse, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, sampleResponse.ID, resp.ID)
				assert.Equal(t, sampleResponse.PVZID, resp.PVZID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.expect()
				resp, err := ctrl.CreateReception(ctx, tt.req)
				tt.assertions(resp, err)
			},
		)
	}
}

func TestController_AddItemToReception(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	repoMock := mocks.NewMockAppRepo(mockCtrl)
	authMock := mocks.NewMockCore(mockCtrl)
	ctrl := New(repoMock, authMock)

	testPVZID := uuid.New()
	testType := "validType"
	baseReq := &dto.AddItemRequest{
		PVZID: testPVZID,
		Type:  testType,
	}

	genericErr := errors.New("generic error")
	sampleResponse := &dto.AddItemResponse{
		ID: testPVZID,
	}

	tests := []struct {
		name       string
		req        *dto.AddItemRequest
		expect     func()
		assertions func(resp *dto.AddItemResponse, err error)
	}{
		{
			name: "No active reception error",
			req:  baseReq,
			expect: func() {
				repoMock.EXPECT().
					AddItemToReception(ctx, gomock.Any()).
					Return(nil, repo.ErrNoActiveReception)
			},
			assertions: func(resp *dto.AddItemResponse, err error) {
				assert.Nil(t, resp)
				assert.ErrorIs(t, err, ErrNoActiveReception)
			},
		},
		{
			name: "Type is not valid error",
			req:  baseReq,
			expect: func() {
				repoMock.EXPECT().
					AddItemToReception(ctx, gomock.Any()).
					Return(nil, repo.ErrTypeIsNotValid)
			},
			assertions: func(resp *dto.AddItemResponse, err error) {
				assert.Nil(t, resp)
				assert.ErrorIs(t, err, ErrTypeIsNotValid)
			},
		},
		{
			name: "General error",
			req:  baseReq,
			expect: func() {
				repoMock.EXPECT().
					AddItemToReception(ctx, gomock.Any()).
					Return(nil, genericErr)
			},
			assertions: func(resp *dto.AddItemResponse, err error) {
				assert.Nil(t, resp)
				assert.Equal(t, genericErr, err)
			},
		},
		{
			name: "Successful addition",
			req:  baseReq,
			expect: func() {
				repoMock.EXPECT().
					AddItemToReception(ctx, &dto.AddItemRequest{PVZID: testPVZID, Type: testType}).
					Return(sampleResponse, nil)
			},
			assertions: func(resp *dto.AddItemResponse, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.expect()
				resp, err := ctrl.AddItemToReception(ctx, tt.req)
				tt.assertions(resp, err)
			},
		)
	}
}

func TestController_GetPVZList(t *testing.T) {
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	repoMock := mocks.NewMockAppRepo(mockCtrl)
	authMock := mocks.NewMockCore(mockCtrl)
	ctrl := New(repoMock, authMock)

	testErr := errors.New("test error")
	samplePVZList := []*md.PVZ{
		{ID: uuid.New()},
		{ID: uuid.New()},
	}

	tests := []struct {
		name       string
		expect     func()
		assertions func(resp []*md.PVZ, err error)
	}{
		{
			name: "Repository error",
			expect: func() {
				repoMock.EXPECT().
					GetPVZList(ctx).
					Return(nil, testErr)
			},
			assertions: func(resp []*md.PVZ, err error) {
				assert.Nil(t, resp)
				assert.Equal(t, testErr, err)
			},
		},
		{
			name: "Successful get PVZ list",
			expect: func() {
				repoMock.EXPECT().
					GetPVZList(ctx).
					Return(samplePVZList, nil)
			},
			assertions: func(resp []*md.PVZ, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Len(t, resp, len(samplePVZList))
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.expect()
				resp, err := ctrl.GetPVZList(ctx)
				tt.assertions(resp, err)
			},
		)
	}
}
