package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/faqears/faqears/services/playlist-service/internal/domain"
)

type PlaylistRepo struct {
	db *pgxpool.Pool
}

func NewPlaylistRepo(db *pgxpool.Pool) *PlaylistRepo {
	return &PlaylistRepo{db: db}
}

func (r *PlaylistRepo) Create(ctx context.Context, p *domain.Playlist) error {
	const q = `
INSERT INTO playlists (id, owner_id, name, description, public, permalink, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Exec(ctx, q,
		p.ID, p.OwnerID, p.Name, p.Description, p.Public, p.Permalink, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *PlaylistRepo) GetByID(ctx context.Context, id string) (*domain.Playlist, error) {
	const q = `
SELECT id, owner_id, name, description, public, permalink, created_at, updated_at
FROM playlists WHERE id = $1`
	p := &domain.Playlist{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&p.ID, &p.OwnerID, &p.Name, &p.Description, &p.Public, &p.Permalink, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PlaylistRepo) GetByPermalink(ctx context.Context, permalink string) (*domain.Playlist, error) {
	const q = `
SELECT id, owner_id, name, description, public, permalink, created_at, updated_at
FROM playlists WHERE permalink = $1`
	p := &domain.Playlist{}
	err := r.db.QueryRow(ctx, q, permalink).Scan(
		&p.ID, &p.OwnerID, &p.Name, &p.Description, &p.Public, &p.Permalink, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PlaylistRepo) Update(ctx context.Context, p *domain.Playlist) error {
	const q = `
UPDATE playlists SET name = $1, description = $2, public = $3, updated_at = $4
WHERE id = $5`
	_, err := r.db.Exec(ctx, q, p.Name, p.Description, p.Public, p.UpdatedAt, p.ID)
	return err
}

func (r *PlaylistRepo) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM playlists WHERE id = $1`
	_, err := r.db.Exec(ctx, q, id)
	return err
}

func (r *PlaylistRepo) ListByOwner(ctx context.Context, ownerID string, limit, offset int32) ([]*domain.Playlist, error) {
	const q = `
SELECT id, owner_id, name, description, public, permalink, created_at, updated_at
FROM playlists WHERE owner_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, q, ownerID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.Playlist
	for rows.Next() {
		p := &domain.Playlist{}
		if err := rows.Scan(
			&p.ID, &p.OwnerID, &p.Name, &p.Description, &p.Public, &p.Permalink, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *PlaylistRepo) ListByCollaborator(ctx context.Context, userID string, limit, offset int32) ([]*domain.Playlist, error) {
	const q = `
SELECT p.id, p.owner_id, p.name, p.description, p.public, p.permalink, p.created_at, p.updated_at
FROM playlists p
JOIN playlist_collaborators pc ON pc.playlist_id = p.id
WHERE pc.user_id = $1
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.Playlist
	for rows.Next() {
		p := &domain.Playlist{}
		if err := rows.Scan(
			&p.ID, &p.OwnerID, &p.Name, &p.Description, &p.Public, &p.Permalink, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *PlaylistRepo) IsCollaborator(ctx context.Context, playlistID, userID string) (bool, error) {
	const q = `SELECT 1 FROM playlist_collaborators WHERE playlist_id = $1 AND user_id = $2`
	rows, err := r.db.Query(ctx, q, playlistID, userID)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), rows.Err()
}

func (r *PlaylistRepo) AddCollaborator(ctx context.Context, c *domain.Collaborator) error {
	const q = `
INSERT INTO playlist_collaborators (playlist_id, user_id, added_at)
VALUES ($1, $2, $3)
ON CONFLICT (playlist_id, user_id) DO NOTHING`
	_, err := r.db.Exec(ctx, q, c.PlaylistID, c.UserID, c.AddedAt)
	return err
}

func (r *PlaylistRepo) RemoveCollaborator(ctx context.Context, playlistID, userID string) error {
	const q = `DELETE FROM playlist_collaborators WHERE playlist_id = $1 AND user_id = $2`
	_, err := r.db.Exec(ctx, q, playlistID, userID)
	return err
}

func (r *PlaylistRepo) GetCollaboratorIDs(ctx context.Context, playlistID string) ([]string, error) {
	const q = `SELECT user_id FROM playlist_collaborators WHERE playlist_id = $1 ORDER BY added_at`
	rows, err := r.db.Query(ctx, q, playlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

var _ = time.Now
