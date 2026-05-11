package grpc

import (
	"context"

	catalogv1 "github.com/faqears/faqears/gen/go/catalog/v1"
	"github.com/faqears/faqears/pkg/errs"
	"github.com/faqears/faqears/services/catalog-service/internal/domain"
	"github.com/faqears/faqears/services/catalog-service/internal/usecase"
)

type Server struct {
	catalogv1.UnimplementedCatalogServiceServer
	catalog *usecase.Catalog
}

func NewServer(catalog *usecase.Catalog) *Server {
	return &Server{catalog: catalog}
}

func (s *Server) GetArtist(ctx context.Context, req *catalogv1.GetArtistRequest) (*catalogv1.GetArtistResponse, error) {
	a, err := s.catalog.GetArtist(ctx, req.GetArtistId())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &catalogv1.GetArtistResponse{Artist: artistToProto(a)}, nil
}

func (s *Server) GetAlbum(ctx context.Context, req *catalogv1.GetAlbumRequest) (*catalogv1.GetAlbumResponse, error) {
	a, err := s.catalog.GetAlbum(ctx, req.GetAlbumId())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &catalogv1.GetAlbumResponse{Album: albumToProto(a)}, nil
}

func (s *Server) GetTrack(ctx context.Context, req *catalogv1.GetTrackRequest) (*catalogv1.GetTrackResponse, error) {
	t, err := s.catalog.GetTrack(ctx, req.GetTrackId())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &catalogv1.GetTrackResponse{Track: trackToProto(t)}, nil
}

func (s *Server) ListTracksByAlbum(ctx context.Context, req *catalogv1.ListTracksByAlbumRequest) (*catalogv1.ListTracksByAlbumResponse, error) {
	tracks, err := s.catalog.ListTracksByAlbum(ctx, req.GetAlbumId(), req.GetLimit(), req.GetOffset())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	out := make([]*catalogv1.Track, len(tracks))
	for i, t := range tracks {
		out[i] = trackToProto(t)
	}
	return &catalogv1.ListTracksByAlbumResponse{Tracks: out}, nil
}

func (s *Server) ListAlbumsByArtist(ctx context.Context, req *catalogv1.ListAlbumsByArtistRequest) (*catalogv1.ListAlbumsByArtistResponse, error) {
	albums, err := s.catalog.ListAlbumsByArtist(ctx, req.GetArtistId(), req.GetLimit(), req.GetOffset())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	out := make([]*catalogv1.Album, len(albums))
	for i, a := range albums {
		out[i] = albumToProto(a)
	}
	return &catalogv1.ListAlbumsByArtistResponse{Albums: out}, nil
}

func (s *Server) Search(ctx context.Context, req *catalogv1.SearchRequest) (*catalogv1.SearchResponse, error) {
	tracks, albums, artists, err := s.catalog.Search(ctx, req.GetQuery(), req.GetLimit())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	protoTracks := make([]*catalogv1.Track, len(tracks))
	for i, t := range tracks {
		protoTracks[i] = trackToProto(t)
	}
	protoAlbums := make([]*catalogv1.Album, len(albums))
	for i, a := range albums {
		protoAlbums[i] = albumToProto(a)
	}
	protoArtists := make([]*catalogv1.Artist, len(artists))
	for i, a := range artists {
		protoArtists[i] = artistToProto(a)
	}
	return &catalogv1.SearchResponse{
		Tracks:  protoTracks,
		Albums:  protoAlbums,
		Artists: protoArtists,
	}, nil
}

func (s *Server) IngestArtist(ctx context.Context, req *catalogv1.IngestArtistRequest) (*catalogv1.IngestArtistResponse, error) {
	id, err := s.catalog.IngestArtist(ctx, req.GetName(), req.GetCountry(), req.GetBiography())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &catalogv1.IngestArtistResponse{ArtistId: id}, nil
}

func (s *Server) IngestAlbum(ctx context.Context, req *catalogv1.IngestAlbumRequest) (*catalogv1.IngestAlbumResponse, error) {
	id, err := s.catalog.IngestAlbum(ctx, req.GetArtistId(), req.GetTitle(), req.GetYear(), req.GetCoverUrl())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &catalogv1.IngestAlbumResponse{AlbumId: id}, nil
}

func (s *Server) IngestTrack(ctx context.Context, req *catalogv1.IngestTrackRequest) (*catalogv1.IngestTrackResponse, error) {
	id, err := s.catalog.IngestTrack(ctx,
		req.GetAlbumId(),
		req.GetArtistId(),
		req.GetTitle(),
		req.GetDurationSec(),
		req.GetIsrc(),
		req.GetGenres(),
		req.GetUserGenerated(),
		req.GetOwnerUserId(),
	)
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &catalogv1.IngestTrackResponse{TrackId: id}, nil
}

func artistToProto(a *domain.Artist) *catalogv1.Artist {
	return &catalogv1.Artist{
		Id:        a.ID,
		Name:      a.Name,
		Country:   a.Country,
		Biography: a.Biography,
		CreatedAt: a.CreatedAt.Unix(),
	}
}

func albumToProto(a *domain.Album) *catalogv1.Album {
	return &catalogv1.Album{
		Id:        a.ID,
		ArtistId:  a.ArtistID,
		Title:     a.Title,
		Year:      a.Year,
		CoverUrl:  a.CoverURL,
		CreatedAt: a.CreatedAt.Unix(),
	}
}

func trackToProto(t *domain.Track) *catalogv1.Track {
	return &catalogv1.Track{
		Id:            t.ID,
		AlbumId:       t.AlbumID,
		ArtistId:      t.ArtistID,
		Title:         t.Title,
		DurationSec:   t.DurationSec,
		Isrc:          t.ISRC,
		Genres:        t.Genres,
		UserGenerated: t.UserGenerated,
		OwnerUserId:   t.OwnerUserID,
		CreatedAt:     t.CreatedAt.Unix(),
	}
}
