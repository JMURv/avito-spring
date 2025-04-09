package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/JMURv/avito-spring/internal/auth"
	"github.com/JMURv/avito-spring/internal/ctrl"
	"github.com/JMURv/avito-spring/internal/dto"
	"github.com/JMURv/avito-spring/internal/hdl"
	"github.com/JMURv/avito-spring/internal/hdl/http/utils"
	md "github.com/JMURv/avito-spring/internal/models"
	"github.com/JMURv/avito-spring/tests/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandler_DummyLogin(t *testing.T) {
	const uri = "/dummyLogin"
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockAppCtrl(mock)
	auth := mocks.NewMockCore(mock)
	h := New(mctrl, auth)

	testErr := errors.New("test-err")

	tests := []struct {
		name       string
		method     string
		status     int
		payload    map[string]any
		expect     func()
		assertions func(r io.ReadCloser)
	}{
		{
			name:   "ErrDecodeRequest",
			method: http.MethodPost,
			status: http.StatusBadRequest,
			payload: map[string]any{
				"role": 0,
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, hdl.ErrDecodeRequest.Error(), res.Message)
			},
			expect: func() {},
		},
		{
			name:   "ErrMissingRole",
			method: http.MethodPost,
			status: http.StatusBadRequest,
			payload: map[string]any{
				"role": "",
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Contains(t, res.Message, "failed on the 'required' tag")
			},
			expect: func() {},
		},
		{
			name:   "StatusInternalServerError",
			method: http.MethodPost,
			status: http.StatusInternalServerError,
			payload: map[string]any{
				"role": "client",
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, res.Message, hdl.ErrInternal.Error())
			},
			expect: func() {
				mctrl.EXPECT().DummyLogin(
					gomock.Any(), &dto.DummyLoginRequest{
						Role: "client",
					},
				).Return(nil, testErr)
			},
		},
		{
			name:   "Success",
			method: http.MethodPost,
			status: http.StatusOK,
			payload: map[string]any{
				"role": "client",
			},
			assertions: func(r io.ReadCloser) {
				res := &dto.DummyLoginResponse{Token: "token"}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, "token", res.Token)
			},
			expect: func() {
				mctrl.EXPECT().DummyLogin(
					gomock.Any(), &dto.DummyLoginRequest{
						Role: "client",
					},
				).Return(&dto.DummyLoginResponse{Token: "token"}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.expect()
				b, err := json.Marshal(tt.payload)
				require.NoError(t, err)

				req := httptest.NewRequest(tt.method, uri, bytes.NewBuffer(b))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				h.dummyLogin(w, req)
				assert.Equal(t, tt.status, w.Result().StatusCode)

				defer func() {
					assert.Nil(t, w.Result().Body.Close())
				}()

				tt.assertions(w.Result().Body)
			},
		)
	}
}

func TestHandler_Register(t *testing.T) {
	const uri = "/register"
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockAppCtrl(mock)
	auth := mocks.NewMockCore(mock)
	h := New(mctrl, auth)

	testErr := errors.New("test-err")

	tests := []struct {
		name       string
		method     string
		status     int
		payload    map[string]any
		expect     func()
		assertions func(r io.ReadCloser)
	}{
		{
			name:   "ErrDecodeRequest",
			method: http.MethodPost,
			status: http.StatusBadRequest,
			payload: map[string]any{
				"email": 123,
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, hdl.ErrDecodeRequest.Error(), res.Message)
			},
			expect: func() {},
		},
		{
			name:   "ValidationError",
			method: http.MethodPost,
			status: http.StatusBadRequest,
			payload: map[string]any{
				"email": "",
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Contains(t, res.Message, "failed on the 'required' tag")
			},
			expect: func() {},
		},
		{
			name:   "InternalError",
			method: http.MethodPost,
			status: http.StatusInternalServerError,
			payload: map[string]any{
				"email":    "test@example.com",
				"role":     "client",
				"password": "password",
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, hdl.ErrInternal.Error(), res.Message)
			},
			expect: func() {
				mctrl.EXPECT().Register(gomock.Any(), gomock.Any()).Return(nil, testErr)
			},
		},
		{
			name:   "Success",
			method: http.MethodPost,
			status: http.StatusCreated,
			payload: map[string]any{
				"email":    "test@example.com",
				"role":     "client",
				"password": "password",
			},
			assertions: func(r io.ReadCloser) {
				res := &dto.RegisterResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.NotNil(t, res)
			},
			expect: func() {
				mctrl.EXPECT().Register(gomock.Any(), gomock.Any()).Return(&dto.RegisterResponse{}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.expect()
				b, _ := json.Marshal(tt.payload)
				req := httptest.NewRequest(tt.method, uri, bytes.NewBuffer(b))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				h.register(w, req)
				assert.Equal(t, tt.status, w.Result().StatusCode)

				defer w.Result().Body.Close()
				tt.assertions(w.Result().Body)
			},
		)
	}
}

func TestHandler_Login(t *testing.T) {
	const uri = "/login"
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockAppCtrl(mock)
	au := mocks.NewMockCore(mock)
	h := New(mctrl, au)

	testErr := errors.New("test-err")

	tests := []struct {
		name       string
		method     string
		status     int
		payload    map[string]any
		expect     func()
		assertions func(r io.ReadCloser)
	}{
		{
			name:   "ErrDecodeRequest",
			method: http.MethodPost,
			status: http.StatusBadRequest,
			payload: map[string]any{
				"email":    123,
				"password": "password",
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, hdl.ErrDecodeRequest.Error(), res.Message)
			},
			expect: func() {},
		},
		{
			name:   "ValidationError",
			method: http.MethodPost,
			status: http.StatusBadRequest,
			payload: map[string]any{
				"email":    "",
				"password": "password",
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Contains(t, res.Message, "failed on the 'required' tag")
			},
			expect: func() {},
		},
		{
			name:   "InternalError",
			method: http.MethodPost,
			status: http.StatusInternalServerError,
			payload: map[string]any{
				"email":    "test@example.com",
				"password": "password",
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, hdl.ErrInternal.Error(), res.Message)
			},
			expect: func() {
				mctrl.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, testErr)
			},
		},
		{
			name:   "InvalidCredentials",
			method: http.MethodPost,
			status: http.StatusUnauthorized,
			payload: map[string]any{
				"email":    "test@example.com",
				"password": "wrong",
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, auth.ErrInvalidCredentials.Error(), res.Message)
			},
			expect: func() {
				mctrl.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, auth.ErrInvalidCredentials)
			},
		},
		{
			name:   "Success",
			method: http.MethodPost,
			status: http.StatusOK,
			payload: map[string]any{
				"email":    "test@example.com",
				"password": "password",
			},
			assertions: func(r io.ReadCloser) {
				res := &dto.LoginResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.NotNil(t, res)
			},
			expect: func() {
				mctrl.EXPECT().Login(gomock.Any(), gomock.Any()).Return(&dto.LoginResponse{}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.expect()
				b, _ := json.Marshal(tt.payload)
				req := httptest.NewRequest(tt.method, uri, bytes.NewBuffer(b))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				h.login(w, req)
				assert.Equal(t, tt.status, w.Result().StatusCode)

				defer w.Result().Body.Close()
				tt.assertions(w.Result().Body)
			},
		)
	}
}

func TestHandler_getPVZ(t *testing.T) {
	const uri = "/pvz"
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockAppCtrl(mock)
	au := mocks.NewMockCore(mock)
	h := New(mctrl, au)

	testErr := errors.New("test error")
	var sampleResponse []*dto.GetPVZResponse

	// Set generic start and end times.
	// For instance, start time 48 hours ago and end time as current time.
	defaultStart := time.Now().Add(-48 * time.Hour).Truncate(time.Second).UTC()
	defaultEnd := time.Now().Truncate(time.Second).UTC()

	// Format times into RFC3339 strings for query parameters.
	startStr := defaultStart.Format(time.RFC3339)
	endStr := defaultEnd.Format(time.RFC3339)

	tests := []struct {
		name        string
		method      string
		status      int
		queryParams map[string]string
		expect      func()
		assertions  func(r io.ReadCloser)
	}{
		{
			name:   "Internal Error",
			method: http.MethodGet,
			status: http.StatusInternalServerError,
			queryParams: map[string]string{
				"page":      "2",
				"limit":     "5",
				"startDate": startStr,
				"endDate":   endStr,
			},
			expect: func() {
				mctrl.EXPECT().
					GetPVZ(gomock.Any(), int64(2), int64(5), defaultStart, defaultEnd).
					Return(nil, testErr)
			},
			assertions: func(r io.ReadCloser) {
				defer r.Close()
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, hdl.ErrInternal.Error(), res.Message)
			},
		},
		{
			name:   "Success",
			method: http.MethodGet,
			status: http.StatusOK,
			queryParams: map[string]string{
				"page":      "3",
				"limit":     "10",
				"startDate": startStr,
				"endDate":   endStr,
			},
			expect: func() {
				mctrl.EXPECT().
					GetPVZ(gomock.Any(), int64(3), int64(10), defaultStart, defaultEnd).
					Return(sampleResponse, nil)
			},
			assertions: func(r io.ReadCloser) {
				defer r.Close()
				var res []*dto.GetPVZResponse
				err := json.NewDecoder(r).Decode(&res)
				assert.Nil(t, err)
				assert.Len(t, res, len(sampleResponse))
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.expect()
				req := httptest.NewRequest(tt.method, uri, nil)
				q := req.URL.Query()
				for k, v := range tt.queryParams {
					q.Add(k, v)
				}
				req.URL.RawQuery = q.Encode()

				w := httptest.NewRecorder()
				h.getPVZ(w, req)
				assert.Equal(t, tt.status, w.Result().StatusCode)

				tt.assertions(w.Result().Body)
			},
		)
	}
}

func TestHandler_CreatePVZ(t *testing.T) {
	const uri = "/pvz"
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockAppCtrl(mock)
	au := mocks.NewMockCore(mock)
	h := New(mctrl, au)

	testErr := errors.New("test-err")

	tests := []struct {
		name       string
		method     string
		status     int
		payload    map[string]any
		expect     func()
		assertions func(r io.ReadCloser)
	}{
		{
			name:   "ErrDecodeRequest",
			method: http.MethodPost,
			status: http.StatusUnauthorized,
			payload: map[string]any{
				"city": 123,
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, hdl.ErrDecodeRequest.Error(), res.Message)
			},
			expect: func() {},
		},
		{
			name:   "ValidationError",
			method: http.MethodPost,
			status: http.StatusUnauthorized,
			payload: map[string]any{
				"city": "",
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Contains(t, res.Message, "failed on the 'required' tag")
			},
			expect: func() {},
		},
		{
			name:   "ErrCityIsNotValid",
			method: http.MethodPost,
			status: http.StatusBadRequest,
			payload: map[string]any{
				"city": "InvalidCity",
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, ctrl.ErrCityIsNotValid.Error(), res.Message)
			},
			expect: func() {
				mctrl.EXPECT().CreatePVZ(gomock.Any(), gomock.Any()).Return(nil, ctrl.ErrCityIsNotValid)
			},
		},
		{
			name:   "InternalError",
			method: http.MethodPost,
			status: http.StatusInternalServerError,
			payload: map[string]any{
				"city": "ValidCity",
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, hdl.ErrInternal.Error(), res.Message)
			},
			expect: func() {
				mctrl.EXPECT().CreatePVZ(gomock.Any(), gomock.Any()).Return(nil, testErr)
			},
		},
		{
			name:   "Success",
			method: http.MethodPost,
			status: http.StatusCreated,
			payload: map[string]any{
				"city": "ValidCity",
			},
			assertions: func(r io.ReadCloser) {
				res := &dto.CreatePVZResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.NotNil(t, res)
			},
			expect: func() {
				mctrl.EXPECT().CreatePVZ(gomock.Any(), gomock.Any()).Return(&dto.CreatePVZResponse{}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.expect()
				b, err := json.Marshal(tt.payload)
				require.NoError(t, err)

				req := httptest.NewRequest(tt.method, uri, bytes.NewBuffer(b))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				h.createPVZ(w, req)
				assert.Equal(t, tt.status, w.Result().StatusCode)

				defer w.Result().Body.Close()
				tt.assertions(w.Result().Body)
			},
		)
	}
}

func TestHandler_CloseLastReception(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockAppCtrl(mock)
	au := mocks.NewMockCore(mock)
	h := New(mctrl, au)

	testErr := errors.New("test-err")

	tests := []struct {
		name       string
		url        string
		status     int
		expect     func()
		assertions func(r io.ReadCloser)
	}{
		{
			name:   "ErrInvalidPathSegments",
			url:    fmt.Sprintf("/pvz/%s/close_last_reception", "wro/ng"),
			status: http.StatusBadRequest,
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, ErrInvalidPathSegments.Error(), res.Message)
			},
			expect: func() {},
		},
		{
			name:   "ErrFailedToParseUUID",
			url:    fmt.Sprintf("/pvz/%s/close_last_reception", uuid.New().String()+"wrong"),
			status: http.StatusBadRequest,
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, ErrFailedToParseUUID.Error(), res.Message)
			},
			expect: func() {},
		},
		{
			name:   "ErrReceptionAlreadyClosed",
			url:    fmt.Sprintf("/pvz/%s/close_last_reception", uuid.New().String()),
			status: http.StatusBadRequest,
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, ctrl.ErrReceptionAlreadyClosed.Error(), res.Message)
			},
			expect: func() {
				mctrl.EXPECT().CloseLastReception(gomock.Any(), gomock.Any()).Return(
					nil,
					ctrl.ErrReceptionAlreadyClosed,
				)
			},
		},
		{
			name:   "InternalError",
			url:    fmt.Sprintf("/pvz/%s/close_last_reception", uuid.New().String()),
			status: http.StatusInternalServerError,
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, hdl.ErrInternal.Error(), res.Message)
			},
			expect: func() {
				mctrl.EXPECT().CloseLastReception(gomock.Any(), gomock.Any()).Return(nil, testErr)
			},
		},
		{
			name:       "Success",
			url:        fmt.Sprintf("/pvz/%s/close_last_reception", uuid.New().String()),
			status:     http.StatusOK,
			assertions: func(r io.ReadCloser) {},
			expect: func() {
				mctrl.EXPECT().CloseLastReception(gomock.Any(), gomock.Any()).Return(&md.Reception{}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.expect()
				req := httptest.NewRequest(http.MethodPost, tt.url, nil)
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				h.closeLastReception(w, req)
				assert.Equal(t, tt.status, w.Result().StatusCode)

				defer w.Result().Body.Close()
				tt.assertions(w.Result().Body)
			},
		)
	}
}

func TestHandler_DeleteLastProduct(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockAppCtrl(mock)
	au := mocks.NewMockCore(mock)
	h := New(mctrl, au)

	testErr := errors.New("test-err")

	tests := []struct {
		name       string
		url        string
		status     int
		expect     func()
		assertions func(r io.ReadCloser)
	}{
		{
			name:   "ErrInvalidPathSegments",
			url:    fmt.Sprintf("/pvz/%s/delete_last_product", "wro/ng"),
			status: http.StatusBadRequest,
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, ErrInvalidPathSegments.Error(), res.Message)
			},
			expect: func() {},
		},
		{
			name:   "ErrFailedToParseUUID",
			url:    fmt.Sprintf("/pvz/%s/delete_last_product", uuid.New().String()+"wrong"),
			status: http.StatusBadRequest,
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, ErrFailedToParseUUID.Error(), res.Message)
			},
			expect: func() {},
		},
		{
			name:   "ErrNoActiveReception",
			url:    fmt.Sprintf("/pvz/%s/delete_last_product", uuid.New().String()),
			status: http.StatusBadRequest,
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, ctrl.ErrNoActiveReception.Error(), res.Message)
			},
			expect: func() {
				mctrl.EXPECT().DeleteLastProduct(gomock.Any(), gomock.Any()).Return(ctrl.ErrNoActiveReception)
			},
		},
		{
			name:   "ErrNoItems",
			url:    fmt.Sprintf("/pvz/%s/delete_last_product", uuid.New().String()),
			status: http.StatusBadRequest,
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, ctrl.ErrNoItems.Error(), res.Message)
			},
			expect: func() {
				mctrl.EXPECT().DeleteLastProduct(gomock.Any(), gomock.Any()).Return(ctrl.ErrNoItems)
			},
		},
		{
			name:   "InternalError",
			url:    fmt.Sprintf("/pvz/%s/delete_last_product", uuid.New().String()),
			status: http.StatusInternalServerError,
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, hdl.ErrInternal.Error(), res.Message)
			},
			expect: func() {
				mctrl.EXPECT().DeleteLastProduct(gomock.Any(), gomock.Any()).Return(testErr)
			},
		},
		{
			name:       "Success",
			url:        fmt.Sprintf("/pvz/%s/delete_last_product", uuid.New().String()),
			status:     http.StatusOK,
			assertions: func(r io.ReadCloser) {},
			expect: func() {
				mctrl.EXPECT().DeleteLastProduct(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.expect()
				req := httptest.NewRequest(http.MethodPost, tt.url, nil)
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				h.deleteLastProduct(w, req)
				assert.Equal(t, tt.status, w.Result().StatusCode)

				defer w.Result().Body.Close()
				tt.assertions(w.Result().Body)
			},
		)
	}
}

func TestHandler_CreateReception(t *testing.T) {
	const uri = "/receptions"
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockAppCtrl(mock)
	auth := mocks.NewMockCore(mock)
	h := New(mctrl, auth)

	testErr := errors.New("test-err")

	tests := []struct {
		name       string
		method     string
		status     int
		payload    map[string]any
		expect     func()
		assertions func(r io.ReadCloser)
	}{
		{
			name:   "ErrDecodeRequest",
			method: http.MethodPost,
			status: http.StatusBadRequest,
			payload: map[string]any{
				"pvzId": 123,
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, hdl.ErrDecodeRequest.Error(), res.Message)
			},
			expect: func() {},
		},
		{
			name:   "ValidationError",
			method: http.MethodPost,
			status: http.StatusBadRequest,
			payload: map[string]any{
				"wrong": "",
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Contains(t, res.Message, "failed on the 'required' tag")
			},
			expect: func() {},
		},
		{
			name:   "ErrReceptionStillOpen",
			method: http.MethodPost,
			status: http.StatusBadRequest,
			payload: map[string]any{
				"pvzId": uuid.New().String(),
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, ctrl.ErrReceptionStillOpen.Error(), res.Message)
			},
			expect: func() {
				mctrl.EXPECT().CreateReception(gomock.Any(), gomock.Any()).Return(nil, ctrl.ErrReceptionStillOpen)
			},
		},
		{
			name:   "InternalError",
			method: http.MethodPost,
			status: http.StatusInternalServerError,
			payload: map[string]any{
				"pvzId": uuid.New().String(),
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, hdl.ErrInternal.Error(), res.Message)
			},
			expect: func() {
				mctrl.EXPECT().CreateReception(gomock.Any(), gomock.Any()).Return(nil, testErr)
			},
		},
		{
			name:   "Success",
			method: http.MethodPost,
			status: http.StatusCreated,
			payload: map[string]any{
				"pvzId":    uuid.New().String(),
				"password": "password",
			},
			assertions: func(r io.ReadCloser) {
				res := &dto.CreateReceptionResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.NotNil(t, res)
			},
			expect: func() {
				mctrl.EXPECT().CreateReception(gomock.Any(), gomock.Any()).Return(&dto.CreateReceptionResponse{}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.expect()
				b, _ := json.Marshal(tt.payload)
				req := httptest.NewRequest(tt.method, uri, bytes.NewBuffer(b))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				h.createReception(w, req)
				assert.Equal(t, tt.status, w.Result().StatusCode)

				defer w.Result().Body.Close()
				tt.assertions(w.Result().Body)
			},
		)
	}
}

func TestHandler_AddItemToReception(t *testing.T) {
	const uri = "/products"
	mock := gomock.NewController(t)
	defer mock.Finish()

	mctrl := mocks.NewMockAppCtrl(mock)
	auth := mocks.NewMockCore(mock)
	h := New(mctrl, auth)

	testErr := errors.New("test-err")

	tests := []struct {
		name       string
		method     string
		status     int
		payload    map[string]any
		expect     func()
		assertions func(r io.ReadCloser)
	}{
		{
			name:   "ErrDecodeRequest",
			method: http.MethodPost,
			status: http.StatusBadRequest,
			payload: map[string]any{
				"type":  "type",
				"pvzId": 123,
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, hdl.ErrDecodeRequest.Error(), res.Message)
			},
			expect: func() {},
		},
		{
			name:   "ValidationError",
			method: http.MethodPost,
			status: http.StatusBadRequest,
			payload: map[string]any{
				"wrong": "",
				"type":  "type",
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Contains(t, res.Message, "failed on the 'required' tag")
			},
			expect: func() {},
		},
		{
			name:   "ErrNoActiveReception",
			method: http.MethodPost,
			status: http.StatusBadRequest,
			payload: map[string]any{
				"pvzId": uuid.New().String(),
				"type":  "type",
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, ctrl.ErrNoActiveReception.Error(), res.Message)
			},
			expect: func() {
				mctrl.EXPECT().AddItemToReception(gomock.Any(), gomock.Any()).Return(nil, ctrl.ErrNoActiveReception)
			},
		},
		{
			name:   "ErrTypeIsNotValid",
			method: http.MethodPost,
			status: http.StatusBadRequest,
			payload: map[string]any{
				"pvzId": uuid.New().String(),
				"type":  "type",
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, ctrl.ErrTypeIsNotValid.Error(), res.Message)
			},
			expect: func() {
				mctrl.EXPECT().AddItemToReception(gomock.Any(), gomock.Any()).Return(nil, ctrl.ErrTypeIsNotValid)
			},
		},
		{
			name:   "InternalError",
			method: http.MethodPost,
			status: http.StatusInternalServerError,
			payload: map[string]any{
				"pvzId": uuid.New().String(),
				"type":  "type",
			},
			assertions: func(r io.ReadCloser) {
				res := &utils.ErrorResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.Equal(t, hdl.ErrInternal.Error(), res.Message)
			},
			expect: func() {
				mctrl.EXPECT().AddItemToReception(gomock.Any(), gomock.Any()).Return(nil, testErr)
			},
		},
		{
			name:   "Success",
			method: http.MethodPost,
			status: http.StatusCreated,
			payload: map[string]any{
				"pvzId":    uuid.New().String(),
				"type":     "type",
				"password": "password",
			},
			assertions: func(r io.ReadCloser) {
				res := &dto.AddItemResponse{}
				err := json.NewDecoder(r).Decode(res)
				assert.Nil(t, err)
				assert.NotNil(t, res)
			},
			expect: func() {
				mctrl.EXPECT().AddItemToReception(gomock.Any(), gomock.Any()).Return(
					&dto.AddItemResponse{},
					nil,
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.expect()
				b, _ := json.Marshal(tt.payload)
				req := httptest.NewRequest(tt.method, uri, bytes.NewBuffer(b))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				h.addItemToReception(w, req)
				assert.Equal(t, tt.status, w.Result().StatusCode)

				defer w.Result().Body.Close()
				tt.assertions(w.Result().Body)
			},
		)
	}
}
