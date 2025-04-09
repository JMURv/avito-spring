package mapper

import (
	gen "github.com/JMURv/avito-spring/api/grpc/v1/gen"
	md "github.com/JMURv/avito-spring/internal/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ListPVZsToProto(req []*md.PVZ) []*gen.PVZ {
	res := make([]*gen.PVZ, len(req))
	for i := 0; i < len(req); i++ {
		res[i] = PVZToProto(req[i])
	}

	return res
}

func PVZToProto(pvz *md.PVZ) *gen.PVZ {
	return &gen.PVZ{
		Id:               pvz.ID.String(),
		RegistrationDate: timestamppb.New(pvz.RegistrationDate),
		City:             pvz.City,
	}
}
