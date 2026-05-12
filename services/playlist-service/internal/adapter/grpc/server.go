package grpc

import (
	"context"

	playlistv1 "github.com/faqears/faqears/gen/go/playlist/v1"
	"github.com/faqears/faqears/pkg/errs"
	"github.com/faqears/faqears/services/playlist-service/internal/domain"
	"github.com/faqears/faqears/services/playlist-service/internal/usecase"
)

type Server struct {
	playlistv1.UnimplementedPlaylistServiceServer
	uc *usecase.Playlist
}

func NewServer(uc *usecase.Playlist) *Server {
	return &Server{uc: uc}
}

func (s *Server) CreatePlaylist(ctx context.Context, req *playlistv1.CreatePlaylistRequest) (*playlistv1.CreatePlaylistResponse, error) {
	p, err := s.uc.CreatePlaylist(ctx, req.GetOwnerId(), req.GetName(), req.GetDescription(), req.GetPublic())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &playlistv1.CreatePlaylistResponse{Playlist: toProto(p)}, nil
}

func (s *Server) GetPlaylist(ctx context.Context, req *playlistv1.GetPlaylistRequest) (*playlistv1.GetPlaylistResponse, error) {
	p, err := s.uc.GetPlaylist(ctx, req.GetPlaylistId())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &playlistv1.GetPlaylistResponse{Playlist: toProto(p)}, nil
}

func (s *Server) GetPlaylistByPermalink(ctx context.Context, req *playlistv1.GetPlaylistByPermalinkRequest) (*playlistv1.GetPlaylistResponse, error) {
	p, err := s.uc.GetPlaylistByPermalink(ctx, req.GetPermalink())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &playlistv1.GetPlaylistResponse{Playlist: toProto(p)}, nil
}

func (s *Server) UpdatePlaylist(ctx context.Context, req *playlistv1.UpdatePlaylistRequest) (*playlistv1.UpdatePlaylistResponse, error) {
	p, err := s.uc.UpdatePlaylist(ctx, req.GetPlaylistId(), "", req.GetName(), req.GetDescription(), req.GetPublic())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &playlistv1.UpdatePlaylistResponse{Playlist: toProto(p)}, nil
}

func (s *Server) DeletePlaylist(ctx context.Context, req *playlistv1.DeletePlaylistRequest) (*playlistv1.DeletePlaylistResponse, error) {
	if err := s.uc.DeletePlaylist(ctx, req.GetPlaylistId(), req.GetCallerId()); err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &playlistv1.DeletePlaylistResponse{}, nil
}

func (s *Server) AddTrack(ctx context.Context, req *playlistv1.AddTrackRequest) (*playlistv1.AddTrackResponse, error) {
	p, err := s.uc.AddTrack(ctx, req.GetPlaylistId(), req.GetTrackId(), req.GetAfterTrackId(), req.GetCallerId())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &playlistv1.AddTrackResponse{Playlist: toProto(p)}, nil
}

func (s *Server) RemoveTrack(ctx context.Context, req *playlistv1.RemoveTrackRequest) (*playlistv1.RemoveTrackResponse, error) {
	p, err := s.uc.RemoveTrack(ctx, req.GetPlaylistId(), req.GetTrackId(), req.GetCallerId())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &playlistv1.RemoveTrackResponse{Playlist: toProto(p)}, nil
}

func (s *Server) ReorderTrack(ctx context.Context, req *playlistv1.ReorderTrackRequest) (*playlistv1.ReorderTrackResponse, error) {
	p, err := s.uc.ReorderTrack(ctx, req.GetPlaylistId(), req.GetTrackId(), req.GetAfterTrackId(), req.GetCallerId())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &playlistv1.ReorderTrackResponse{Playlist: toProto(p)}, nil
}

func (s *Server) ListUserPlaylists(ctx context.Context, req *playlistv1.ListUserPlaylistsRequest) (*playlistv1.ListUserPlaylistsResponse, error) {
	ps, err := s.uc.ListUserPlaylists(ctx, req.GetUserId(), req.GetLimit(), req.GetOffset(), req.GetIncludeCollaborations())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	out := make([]*playlistv1.Playlist, len(ps))
	for i, p := range ps {
		out[i] = toProto(p)
	}
	return &playlistv1.ListUserPlaylistsResponse{Playlists: out}, nil
}

func (s *Server) AddCollaborator(ctx context.Context, req *playlistv1.AddCollaboratorRequest) (*playlistv1.AddCollaboratorResponse, error) {
	if err := s.uc.AddCollaborator(ctx, req.GetPlaylistId(), req.GetCollaboratorId(), req.GetCallerId()); err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &playlistv1.AddCollaboratorResponse{}, nil
}

func (s *Server) RemoveCollaborator(ctx context.Context, req *playlistv1.RemoveCollaboratorRequest) (*playlistv1.RemoveCollaboratorResponse, error) {
	if err := s.uc.RemoveCollaborator(ctx, req.GetPlaylistId(), req.GetCollaboratorId(), req.GetCallerId()); err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &playlistv1.RemoveCollaboratorResponse{}, nil
}

func toProto(p *domain.Playlist) *playlistv1.Playlist {
	tracks := make([]*playlistv1.PlaylistTrack, len(p.Tracks))
	for i, t := range p.Tracks {
		tracks[i] = &playlistv1.PlaylistTrack{
			TrackId: t.TrackID,
			Rank:    t.Rank,
			AddedAt: t.AddedAt.Unix(),
			AddedBy: t.AddedBy,
		}
	}
	return &playlistv1.Playlist{
		Id:              p.ID,
		OwnerId:         p.OwnerID,
		Name:            p.Name,
		Description:     p.Description,
		Public:          p.Public,
		Permalink:       p.Permalink,
		CreatedAt:       p.CreatedAt.Unix(),
		UpdatedAt:       p.UpdatedAt.Unix(),
		CollaboratorIds: p.CollaboratorIDs,
		Tracks:          tracks,
	}
}
