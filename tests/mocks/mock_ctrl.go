// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/ctrl/ctrl.go
//
// Generated by this command:
//
//	mockgen -source=./internal/ctrl/ctrl.go -destination=tests/mocks/mock_ctrl.go -package=mocks
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"
	time "time"

	dto "github.com/JMURv/avito-spring/internal/dto"
	models "github.com/JMURv/avito-spring/internal/models"
	uuid "github.com/google/uuid"
	gomock "go.uber.org/mock/gomock"
)

// MockAppRepo is a mock of AppRepo interface.
type MockAppRepo struct {
	ctrl     *gomock.Controller
	recorder *MockAppRepoMockRecorder
	isgomock struct{}
}

// MockAppRepoMockRecorder is the mock recorder for MockAppRepo.
type MockAppRepoMockRecorder struct {
	mock *MockAppRepo
}

// NewMockAppRepo creates a new mock instance.
func NewMockAppRepo(ctrl *gomock.Controller) *MockAppRepo {
	mock := &MockAppRepo{ctrl: ctrl}
	mock.recorder = &MockAppRepoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAppRepo) EXPECT() *MockAppRepoMockRecorder {
	return m.recorder
}

// AddItemToReception mocks base method.
func (m *MockAppRepo) AddItemToReception(ctx context.Context, req *dto.AddItemRequest) (*dto.AddItemResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddItemToReception", ctx, req)
	ret0, _ := ret[0].(*dto.AddItemResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddItemToReception indicates an expected call of AddItemToReception.
func (mr *MockAppRepoMockRecorder) AddItemToReception(ctx, req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddItemToReception", reflect.TypeOf((*MockAppRepo)(nil).AddItemToReception), ctx, req)
}

// CloseLastReception mocks base method.
func (m *MockAppRepo) CloseLastReception(ctx context.Context, id uuid.UUID) (*models.Reception, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CloseLastReception", ctx, id)
	ret0, _ := ret[0].(*models.Reception)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CloseLastReception indicates an expected call of CloseLastReception.
func (mr *MockAppRepoMockRecorder) CloseLastReception(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CloseLastReception", reflect.TypeOf((*MockAppRepo)(nil).CloseLastReception), ctx, id)
}

// CreatePVZ mocks base method.
func (m *MockAppRepo) CreatePVZ(ctx context.Context, req *dto.CreatePVZRequest) (uuid.UUID, time.Time, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreatePVZ", ctx, req)
	ret0, _ := ret[0].(uuid.UUID)
	ret1, _ := ret[1].(time.Time)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CreatePVZ indicates an expected call of CreatePVZ.
func (mr *MockAppRepoMockRecorder) CreatePVZ(ctx, req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreatePVZ", reflect.TypeOf((*MockAppRepo)(nil).CreatePVZ), ctx, req)
}

// CreateReception mocks base method.
func (m *MockAppRepo) CreateReception(ctx context.Context, req *dto.CreateReceptionRequest) (*dto.CreateReceptionResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateReception", ctx, req)
	ret0, _ := ret[0].(*dto.CreateReceptionResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateReception indicates an expected call of CreateReception.
func (mr *MockAppRepoMockRecorder) CreateReception(ctx, req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateReception", reflect.TypeOf((*MockAppRepo)(nil).CreateReception), ctx, req)
}

// CreateUser mocks base method.
func (m *MockAppRepo) CreateUser(ctx context.Context, req *dto.RegisterRequest) (uuid.UUID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUser", ctx, req)
	ret0, _ := ret[0].(uuid.UUID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateUser indicates an expected call of CreateUser.
func (mr *MockAppRepoMockRecorder) CreateUser(ctx, req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUser", reflect.TypeOf((*MockAppRepo)(nil).CreateUser), ctx, req)
}

// DeleteLastProduct mocks base method.
func (m *MockAppRepo) DeleteLastProduct(ctx context.Context, id uuid.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteLastProduct", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteLastProduct indicates an expected call of DeleteLastProduct.
func (mr *MockAppRepoMockRecorder) DeleteLastProduct(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteLastProduct", reflect.TypeOf((*MockAppRepo)(nil).DeleteLastProduct), ctx, id)
}

// GetPVZ mocks base method.
func (m *MockAppRepo) GetPVZ(ctx context.Context, page, limit int64, startDate, endDate time.Time) ([]*dto.GetPVZResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPVZ", ctx, page, limit, startDate, endDate)
	ret0, _ := ret[0].([]*dto.GetPVZResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPVZ indicates an expected call of GetPVZ.
func (mr *MockAppRepoMockRecorder) GetPVZ(ctx, page, limit, startDate, endDate any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPVZ", reflect.TypeOf((*MockAppRepo)(nil).GetPVZ), ctx, page, limit, startDate, endDate)
}

// GetPVZList mocks base method.
func (m *MockAppRepo) GetPVZList(ctx context.Context) ([]*models.PVZ, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPVZList", ctx)
	ret0, _ := ret[0].([]*models.PVZ)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPVZList indicates an expected call of GetPVZList.
func (mr *MockAppRepoMockRecorder) GetPVZList(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPVZList", reflect.TypeOf((*MockAppRepo)(nil).GetPVZList), ctx)
}

// GetUserByEmail mocks base method.
func (m *MockAppRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByEmail", ctx, email)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserByEmail indicates an expected call of GetUserByEmail.
func (mr *MockAppRepoMockRecorder) GetUserByEmail(ctx, email any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByEmail", reflect.TypeOf((*MockAppRepo)(nil).GetUserByEmail), ctx, email)
}

// MockAppCtrl is a mock of AppCtrl interface.
type MockAppCtrl struct {
	ctrl     *gomock.Controller
	recorder *MockAppCtrlMockRecorder
	isgomock struct{}
}

// MockAppCtrlMockRecorder is the mock recorder for MockAppCtrl.
type MockAppCtrlMockRecorder struct {
	mock *MockAppCtrl
}

// NewMockAppCtrl creates a new mock instance.
func NewMockAppCtrl(ctrl *gomock.Controller) *MockAppCtrl {
	mock := &MockAppCtrl{ctrl: ctrl}
	mock.recorder = &MockAppCtrlMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAppCtrl) EXPECT() *MockAppCtrlMockRecorder {
	return m.recorder
}

// AddItemToReception mocks base method.
func (m *MockAppCtrl) AddItemToReception(ctx context.Context, req *dto.AddItemRequest) (*dto.AddItemResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddItemToReception", ctx, req)
	ret0, _ := ret[0].(*dto.AddItemResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddItemToReception indicates an expected call of AddItemToReception.
func (mr *MockAppCtrlMockRecorder) AddItemToReception(ctx, req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddItemToReception", reflect.TypeOf((*MockAppCtrl)(nil).AddItemToReception), ctx, req)
}

// CloseLastReception mocks base method.
func (m *MockAppCtrl) CloseLastReception(ctx context.Context, id uuid.UUID) (*models.Reception, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CloseLastReception", ctx, id)
	ret0, _ := ret[0].(*models.Reception)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CloseLastReception indicates an expected call of CloseLastReception.
func (mr *MockAppCtrlMockRecorder) CloseLastReception(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CloseLastReception", reflect.TypeOf((*MockAppCtrl)(nil).CloseLastReception), ctx, id)
}

// CreatePVZ mocks base method.
func (m *MockAppCtrl) CreatePVZ(ctx context.Context, req *dto.CreatePVZRequest) (*dto.CreatePVZResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreatePVZ", ctx, req)
	ret0, _ := ret[0].(*dto.CreatePVZResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreatePVZ indicates an expected call of CreatePVZ.
func (mr *MockAppCtrlMockRecorder) CreatePVZ(ctx, req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreatePVZ", reflect.TypeOf((*MockAppCtrl)(nil).CreatePVZ), ctx, req)
}

// CreateReception mocks base method.
func (m *MockAppCtrl) CreateReception(ctx context.Context, req *dto.CreateReceptionRequest) (*dto.CreateReceptionResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateReception", ctx, req)
	ret0, _ := ret[0].(*dto.CreateReceptionResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateReception indicates an expected call of CreateReception.
func (mr *MockAppCtrlMockRecorder) CreateReception(ctx, req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateReception", reflect.TypeOf((*MockAppCtrl)(nil).CreateReception), ctx, req)
}

// DeleteLastProduct mocks base method.
func (m *MockAppCtrl) DeleteLastProduct(ctx context.Context, id uuid.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteLastProduct", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteLastProduct indicates an expected call of DeleteLastProduct.
func (mr *MockAppCtrlMockRecorder) DeleteLastProduct(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteLastProduct", reflect.TypeOf((*MockAppCtrl)(nil).DeleteLastProduct), ctx, id)
}

// DummyLogin mocks base method.
func (m *MockAppCtrl) DummyLogin(ctx context.Context, req *dto.DummyLoginRequest) (*dto.DummyLoginResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DummyLogin", ctx, req)
	ret0, _ := ret[0].(*dto.DummyLoginResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DummyLogin indicates an expected call of DummyLogin.
func (mr *MockAppCtrlMockRecorder) DummyLogin(ctx, req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DummyLogin", reflect.TypeOf((*MockAppCtrl)(nil).DummyLogin), ctx, req)
}

// GetPVZ mocks base method.
func (m *MockAppCtrl) GetPVZ(ctx context.Context, page, limit int64, startDate, endDate time.Time) ([]*dto.GetPVZResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPVZ", ctx, page, limit, startDate, endDate)
	ret0, _ := ret[0].([]*dto.GetPVZResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPVZ indicates an expected call of GetPVZ.
func (mr *MockAppCtrlMockRecorder) GetPVZ(ctx, page, limit, startDate, endDate any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPVZ", reflect.TypeOf((*MockAppCtrl)(nil).GetPVZ), ctx, page, limit, startDate, endDate)
}

// GetPVZList mocks base method.
func (m *MockAppCtrl) GetPVZList(ctx context.Context) ([]*models.PVZ, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPVZList", ctx)
	ret0, _ := ret[0].([]*models.PVZ)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPVZList indicates an expected call of GetPVZList.
func (mr *MockAppCtrlMockRecorder) GetPVZList(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPVZList", reflect.TypeOf((*MockAppCtrl)(nil).GetPVZList), ctx)
}

// Login mocks base method.
func (m *MockAppCtrl) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Login", ctx, req)
	ret0, _ := ret[0].(*dto.LoginResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Login indicates an expected call of Login.
func (mr *MockAppCtrlMockRecorder) Login(ctx, req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Login", reflect.TypeOf((*MockAppCtrl)(nil).Login), ctx, req)
}

// Register mocks base method.
func (m *MockAppCtrl) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register", ctx, req)
	ret0, _ := ret[0].(*dto.RegisterResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Register indicates an expected call of Register.
func (mr *MockAppCtrlMockRecorder) Register(ctx, req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockAppCtrl)(nil).Register), ctx, req)
}
