package promotiongrpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"marketplace/internal/promotion/application/commands"
	"marketplace/internal/promotion/application/queries"
	"marketplace/internal/promotion/domain/entities"
	"marketplace/internal/promotion/grpc/greetpb"
)

type GreetServer struct {
	greetpb.UnimplementedGreetServiceServer

	findOne    *queries.FindByCatalogItemHandler
	findMany   *queries.FindManyByCatalogItemsHandler
	addCmd     *commands.AddDiscountHandler
	deactivate *commands.DeactivateDiscountHandler
}

func NewGreetServer(
	findOne *queries.FindByCatalogItemHandler,
	findMany *queries.FindManyByCatalogItemsHandler,
	addCmd *commands.AddDiscountHandler,
	deactivate *commands.DeactivateDiscountHandler,
) *GreetServer {
	return &GreetServer{
		findOne:    findOne,
		findMany:   findMany,
		addCmd:     addCmd,
		deactivate: deactivate,
	}
}

func (s *GreetServer) SayHello(_ context.Context, req *greetpb.HelloRequest) (*greetpb.HelloResponse, error) {
	return &greetpb.HelloResponse{
		Message: fmt.Sprintf("Hello, %s!", req.Name),
	}, nil
}

func (s *GreetServer) GetDiscount(ctx context.Context, req *greetpb.GetDiscountRequest) (*greetpb.GetDiscountResponse, error) {
	if req.ItemId == "" {
		return nil, status.Error(codes.InvalidArgument, "item_id is required")
	}

	discount, err := s.findOne.Handle(ctx, req.ItemId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get discount: %v", err)
	}

	if discount == nil {
		return &greetpb.GetDiscountResponse{
			ItemId:  req.ItemId,
			Percent: 0,
			Active:  false,
		}, nil
	}

	return discountToProto(discount), nil
}

func (s *GreetServer) GetDiscounts(ctx context.Context, req *greetpb.GetDiscountsRequest) (*greetpb.GetDiscountsResponse, error) {
	if len(req.ItemIds) == 0 {
		return &greetpb.GetDiscountsResponse{Discounts: nil}, nil
	}

	discounts, err := s.findMany.Handle(ctx, req.ItemIds)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get discounts: %v", err)
	}

	byItemID := make(map[string]*entities.Discount, len(discounts))
	for i := range discounts {
		byItemID[discounts[i].ItemID] = &discounts[i]
	}

	result := make([]*greetpb.GetDiscountResponse, 0, len(req.ItemIds))
	for _, id := range req.ItemIds {
		if d, ok := byItemID[id]; ok {
			result = append(result, discountToProto(d))
		} else {
			result = append(result, &greetpb.GetDiscountResponse{
				ItemId:  id,
				Percent: 0,
				Active:  false,
			})
		}
	}

	return &greetpb.GetDiscountsResponse{Discounts: result}, nil
}

func (s *GreetServer) AddDiscount(ctx context.Context, req *greetpb.AddDiscountRequest) (*greetpb.AddDiscountResponse, error) {
	if req.ItemId == "" {
		return nil, status.Error(codes.InvalidArgument, "item_id is required")
	}
	if req.Percent < 0 || req.Percent > 100 {
		return nil, status.Errorf(codes.InvalidArgument, "percent must be between 0 and 100, got %.2f", req.Percent)
	}

	cmd := commands.AddDiscountCommand{
		ItemID:  req.ItemId,
		Percent: req.Percent,
	}

	if req.StartsAt != "" {
		t, err := time.Parse(time.RFC3339, req.StartsAt)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid starts_at format, use RFC3339 (e.g. 2024-01-01T00:00:00Z): %v", err)
		}
		cmd.StartsAt = strPtr(t.Format(time.RFC3339))
	}
	if req.EndsAt != "" {
		t, err := time.Parse(time.RFC3339, req.EndsAt)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid ends_at format, use RFC3339 (e.g. 2024-12-31T23:59:59Z): %v", err)
		}
		cmd.EndsAt = strPtr(t.Format(time.RFC3339))
	}

	discountID, err := s.addCmd.Handle(ctx, cmd)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "add discount: %v", err)
	}

	return &greetpb.AddDiscountResponse{
		DiscountId: discountID,
		Ok:         true,
	}, nil
}

func (s *GreetServer) DeactivateDiscount(ctx context.Context, req *greetpb.DeactivateDiscountRequest) (*greetpb.DeactivateDiscountResponse, error) {
	if req.DiscountId == "" {
		return nil, status.Error(codes.InvalidArgument, "discount_id is required")
	}

	err := s.deactivate.Handle(ctx, commands.DeactivateDiscountCommand{
		DiscountID: req.DiscountId,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "deactivate discount: %v", err)
	}

	return &greetpb.DeactivateDiscountResponse{Ok: true}, nil
}

func discountToProto(d *entities.Discount) *greetpb.GetDiscountResponse {
	return &greetpb.GetDiscountResponse{
		ItemId:  d.ItemID,
		Percent: d.Percent,
		Active:  d.Active,
	}
}

func strPtr(s string) *string { return &s }
