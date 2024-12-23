package service

import (
	"context"
	"errors"
	"strings"

	"effective-mobile-go/internal/model"
)

type LocalRepo interface {
	CreateSong(context.Context, model.SongDetail) (model.SongDetail, error)
	ListSongs(context.Context, model.SongFilters) ([]model.SongDetail, error)
	GetSong(_ context.Context, songID uint64) (model.SongDetail, error)
	UpdateSong(context.Context, model.SongUpdate) (model.SongDetail, error)
	DeleteSong(_ context.Context, songID uint64) error
}

type RemoteRepo interface {
	GetSong(context.Context, model.SongDetail) (model.SongDetail, error)
}

type Service struct {
	localRepo  LocalRepo
	remoteRepo RemoteRepo
}

func New(localRepo LocalRepo, remoteRepo RemoteRepo) Service {
	return Service{
		localRepo:  localRepo,
		remoteRepo: remoteRepo,
	}
}

func (s Service) CreateSong(ctx context.Context, song model.SongDetail) (model.SongDetail, error) {
	var zero model.SongDetail

	list, err := s.localRepo.ListSongs(ctx, model.SongFilters{
		Name:  &song.Name,
		Group: &song.Group,
	})

	if err != nil && !errors.Is(err, model.ErrNotFound) {
		return zero, err
	}

	if len(list) > 0 {
		return list[0], nil
	}

	song, err = s.remoteRepo.GetSong(ctx, song)
	if err != nil {
		return zero, err
	}

	return s.localRepo.CreateSong(ctx, song)
}

func (s Service) ListSongs(ctx context.Context, req model.SongFilters) ([]model.SongDetail, error) {
	return s.localRepo.ListSongs(ctx, req)
}

func (s Service) GetSong(ctx context.Context, id uint64) (model.SongDetail, error) {

	song, err := s.localRepo.GetSong(ctx, id)
	if err != nil {
		return song, err
	}

	song.Text = "" // xxx

	return song, nil
}

func (s Service) GetSongText(ctx context.Context, req model.GetSongTextRequest) (verses []string, _ error) {

	song, err := s.localRepo.GetSong(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	verses = strings.Split(strings.TrimSpace(song.Text), "\n\n")

	if req.Offset != nil {
		offset := min(*req.Offset, uint64(len(verses)))
		verses = verses[offset:]
	}

	if req.Limit != nil {
		limit := min(*req.Limit, uint64(len(verses)))
		verses = verses[:limit]
	}

	return verses, nil
}

func (s Service) UpdateSong(ctx context.Context, req model.SongUpdate) (model.SongDetail, error) {
	return s.localRepo.UpdateSong(ctx, req)
}

func (s Service) DeleteSong(ctx context.Context, id uint64) error {
	return s.localRepo.DeleteSong(ctx, id)
}
