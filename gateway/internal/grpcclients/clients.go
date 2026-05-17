package grpcclients

import (
	authv1 "github.com/faqears/faqears/gen/go/auth/v1"
	catalogv1 "github.com/faqears/faqears/gen/go/catalog/v1"
	generationv1 "github.com/faqears/faqears/gen/go/generation/v1"
	paymentv1 "github.com/faqears/faqears/gen/go/payment/v1"
	playlistv1 "github.com/faqears/faqears/gen/go/playlist/v1"
	recommendationv1 "github.com/faqears/faqears/gen/go/recommendation/v1"
	streamingv1 "github.com/faqears/faqears/gen/go/streaming/v1"
	userv1 "github.com/faqears/faqears/gen/go/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Clients struct {
	Auth           authv1.AuthServiceClient
	User           userv1.UserServiceClient
	Catalog        catalogv1.CatalogServiceClient
	Streaming      streamingv1.StreamingServiceClient
	Playlist       playlistv1.PlaylistServiceClient
	Recommendation recommendationv1.RecommendationServiceClient
	Payment        paymentv1.PaymentServiceClient
	Generation     generationv1.GenerationServiceClient

	authConn           *grpc.ClientConn
	userConn           *grpc.ClientConn
	catalogConn        *grpc.ClientConn
	streamingConn      *grpc.ClientConn
	playlistConn       *grpc.ClientConn
	recommendationConn *grpc.ClientConn
	paymentConn        *grpc.ClientConn
	generationConn     *grpc.ClientConn
}

const maxGRPCMessageSize = 64 << 20

func New(authAddr, userAddr, catalogAddr, streamingAddr, playlistAddr, recommendationAddr, paymentAddr, generationAddr string) (*Clients, error) {
	dial := func(addr string) (*grpc.ClientConn, error) {
		return grpc.NewClient(addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultCallOptions(
				grpc.MaxCallSendMsgSize(maxGRPCMessageSize),
				grpc.MaxCallRecvMsgSize(maxGRPCMessageSize),
			),
		)
	}

	authConn, err := dial(authAddr)
	if err != nil {
		return nil, err
	}

	userConn, err := dial(userAddr)
	if err != nil {
		authConn.Close()
		return nil, err
	}

	catalogConn, err := dial(catalogAddr)
	if err != nil {
		authConn.Close()
		userConn.Close()
		return nil, err
	}

	streamingConn, err := dial(streamingAddr)
	if err != nil {
		authConn.Close()
		userConn.Close()
		catalogConn.Close()
		return nil, err
	}

	playlistConn, err := dial(playlistAddr)
	if err != nil {
		authConn.Close()
		userConn.Close()
		catalogConn.Close()
		streamingConn.Close()
		return nil, err
	}

	recommendationConn, err := dial(recommendationAddr)
	if err != nil {
		authConn.Close()
		userConn.Close()
		catalogConn.Close()
		streamingConn.Close()
		playlistConn.Close()
		return nil, err
	}

	paymentConn, err := dial(paymentAddr)
	if err != nil {
		authConn.Close()
		userConn.Close()
		catalogConn.Close()
		streamingConn.Close()
		playlistConn.Close()
		recommendationConn.Close()
		return nil, err
	}

	generationConn, err := dial(generationAddr)
	if err != nil {
		authConn.Close()
		userConn.Close()
		catalogConn.Close()
		streamingConn.Close()
		playlistConn.Close()
		recommendationConn.Close()
		paymentConn.Close()
		return nil, err
	}

	return &Clients{
		Auth:               authv1.NewAuthServiceClient(authConn),
		User:               userv1.NewUserServiceClient(userConn),
		Catalog:            catalogv1.NewCatalogServiceClient(catalogConn),
		Streaming:          streamingv1.NewStreamingServiceClient(streamingConn),
		Playlist:           playlistv1.NewPlaylistServiceClient(playlistConn),
		Recommendation:     recommendationv1.NewRecommendationServiceClient(recommendationConn),
		Payment:            paymentv1.NewPaymentServiceClient(paymentConn),
		Generation:         generationv1.NewGenerationServiceClient(generationConn),
		authConn:           authConn,
		userConn:           userConn,
		catalogConn:        catalogConn,
		streamingConn:      streamingConn,
		playlistConn:       playlistConn,
		recommendationConn: recommendationConn,
		paymentConn:        paymentConn,
		generationConn:     generationConn,
	}, nil
}

func (c *Clients) Close() {
	c.authConn.Close()
	c.userConn.Close()
	c.catalogConn.Close()
	c.streamingConn.Close()
	c.playlistConn.Close()
	c.recommendationConn.Close()
	c.paymentConn.Close()
	c.generationConn.Close()
}
