package controller

import (
	"context"

	pb "go-wordpress/api/proto/product/v1"
	"go-wordpress/internal/product/service"
)

type ProductGRPC struct {
	pb.UnimplementedProductServiceServer
	svc *service.Product
}

func NewGRPC(svc *service.Product) pb.ProductServiceServer {
	return &ProductGRPC{
		svc: svc,
	}
}

func (h *ProductGRPC) GetProductByID(ctx context.Context, req *pb.ProductRequest) (*pb.ProductResponse, error) {
	p, err := h.svc.GetProductByID(ctx, req.Id)
	return &pb.ProductResponse{
		Id: p.ID,
	}, err
}
