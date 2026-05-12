package grpcclient

import (
	"context"

	userv1 "github.com/faqears/faqears/gen/go/user/v1"
	"github.com/faqears/faqears/pkg/errs"
	"github.com/faqears/faqears/pkg/grpcx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserClient struct {
	client userv1.UserServiceClient
}

func NewUserClient(addr string) (*UserClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &UserClient{client: userv1.NewUserServiceClient(conn)}, nil
}

func (c *UserClient) IsPremium(ctx context.Context, userID string) (bool, error) {
	_, err := c.client.GetUser(ctx, &userv1.GetUserRequest{UserId: userID})
	if err != nil {
		return false, errs.Wrap(errs.KindUnavailable, "get user failed", err)
	}
	for _, r := range grpcx.RolesFromCtx(ctx) {
		if r == "premium" {
			return true, nil
		}
	}
	return false, nil
}
