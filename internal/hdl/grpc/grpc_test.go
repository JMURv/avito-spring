package grpc

import (
	"context"
	"errors"
	gen "github.com/JMURv/avito-spring/api/grpc/v1/gen"
	"github.com/JMURv/avito-spring/internal/hdl"
	md "github.com/JMURv/avito-spring/internal/models"
	"github.com/JMURv/avito-spring/tests/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
	"time"
)

func TestHandler_GetPVZList(t *testing.T) {
	mock := gomock.NewController(t)
	defer mock.Finish()

	testErr := errors.New("test error")
	mctrl := mocks.NewMockAppCtrl(mock)
	h := New("test-svc", mctrl)

	tests := []struct {
		name       string
		req        *gen.GetPVZListRequest
		expect     func()
		assertions func(*gen.GetPVZListResponse, error)
	}{
		{
			name:   "NilRequest",
			req:    nil,
			expect: func() {},
			assertions: func(res *gen.GetPVZListResponse, err error) {
				assert.Nil(t, res)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, codes.InvalidArgument, st.Code())
				assert.Equal(t, hdl.ErrDecodeRequest.Error(), st.Message())
			},
		},
		{
			name: "InternalError",
			req:  &gen.GetPVZListRequest{},
			expect: func() {
				mctrl.EXPECT().
					GetPVZList(gomock.Any()).
					Return(nil, testErr)
			},
			assertions: func(res *gen.GetPVZListResponse, err error) {
				assert.Nil(t, res)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, codes.Internal, st.Code())
				assert.Equal(t, hdl.ErrInternal.Error(), st.Message())
			},
		},
		{
			name: "Success",
			req:  &gen.GetPVZListRequest{},
			expect: func() {
				mctrl.EXPECT().
					GetPVZList(gomock.Any()).
					Return(
						[]*md.PVZ{
							{
								ID:               uuid.New(),
								City:             "TestCity",
								RegistrationDate: time.Now(),
							},
						}, nil,
					)
			},
			assertions: func(res *gen.GetPVZListResponse, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Len(t, res.Pvzs, 1)
				assert.Equal(t, "TestCity", res.Pvzs[0].City)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				tt.expect()
				res, err := h.GetPVZList(context.Background(), tt.req)
				tt.assertions(res, err)
			},
		)
	}
}
