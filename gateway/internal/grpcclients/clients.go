package grpcclients

import (
	authv1 "github.com/faqears/faqears/gen/go/auth/v1"
	catalogv1 "github.com/faqears/faqears/gen/go/catalog/v1"
	userv1 "github.com/faqears/faqears/gen/go/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Clients struct {
	Auth    authv1.AuthServiceClient
	User    userv1.UserServiceClient
	Catalog catalogv1.CatalogServiceClient

	authConn    *grpc.ClientConn
	userConn    *grpc.ClientConn
	catalogConn *grpc.ClientConn
}

func New(authAddr, userAddr, catalogAddr string) (*Clients, error) {
	authConn, err := grpc.NewClient(authAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	userConn, err := grpc.NewClient(userAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		authConn.Close()
		return nil, err
	}

	catalogConn, err := grpc.NewClient(catalogAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		authConn.Close()
		userConn.Close()
		return nil, err
	}

	return &Clients{
		Auth:        authv1.NewAuthServiceClient(authConn),
		User:        userv1.NewUserServiceClient(userConn),
		Catalog:     catalogv1.NewCatalogServiceClient(catalogConn),
		authConn:    authConn,
		userConn:    userConn,
		catalogConn: catalogConn,
	}, nil
}

func (c *Clients) Close() {
	c.authConn.Close()
	c.userConn.Close()
	c.catalogConn.Close()
}
