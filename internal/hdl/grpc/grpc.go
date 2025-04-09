package grpc

import (
	"context"
	"errors"
	"fmt"
	gen "github.com/JMURv/avito-spring/api/grpc/v1/gen"
	"github.com/JMURv/avito-spring/internal/ctrl"
	"github.com/JMURv/avito-spring/internal/hdl"
	"github.com/JMURv/avito-spring/internal/models/mapper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"net"
)

type Handler struct {
	gen.PVZServiceServer
	srv  *grpc.Server
	hsrv *health.Server
	ctrl ctrl.AppCtrl
}

func New(name string, ctrl ctrl.AppCtrl) *Handler {
	srv := grpc.NewServer()
	reflection.Register(srv)

	hsrv := health.NewServer()
	hsrv.SetServingStatus(name, grpc_health_v1.HealthCheckResponse_SERVING)
	return &Handler{
		ctrl: ctrl,
		srv:  srv,
		hsrv: hsrv,
	}
}

func (h *Handler) Start(port int) {
	gen.RegisterPVZServiceServer(h.srv, h)
	grpc_health_v1.RegisterHealthServer(h.srv, h.hsrv)

	portStr := fmt.Sprintf(":%v", port)
	lis, err := net.Listen("tcp", portStr)
	if err != nil {
		zap.L().Fatal("failed to listen", zap.Error(err))
	}

	zap.L().Info(
		"Starting GRPC server",
		zap.String("addr", portStr),
	)
	if err = h.srv.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		zap.L().Fatal("failed to serve", zap.Error(err))
	}
}

func (h *Handler) Close(_ context.Context) error {
	h.srv.GracefulStop()
	return nil
}

func (h *Handler) GetPVZList(ctx context.Context, req *gen.GetPVZListRequest) (*gen.GetPVZListResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, hdl.ErrDecodeRequest.Error())
	}

	res, err := h.ctrl.GetPVZList(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, hdl.ErrInternal.Error())
	}

	return &gen.GetPVZListResponse{
		Pvzs: mapper.ListPVZsToProto(res),
	}, nil
}
