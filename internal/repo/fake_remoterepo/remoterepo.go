package fake_remoterepo

import (
	"context"

	"effective-mobile-go/internal/config"
	"effective-mobile-go/internal/logger"
	"effective-mobile-go/internal/model"
)

const loogerGroup = "fake_remoterepo"

type SongDetail = model.SongDetail

type RemoteRepo struct{}

func New(_ config.RemoteAPI) RemoteRepo { return RemoteRepo{} }

func (r RemoteRepo) GetSong(ctx context.Context, song SongDetail) (SongDetail, error) {
	const op = "GetSong"
	var zero SongDetail

	if song.Name == "Supermassive Black Hole" && song.Group == "Muse" {

		release, _ := model.ParseDate("16.07.2006")

		return SongDetail{
			ID:      0,
			Name:    "Supermassive Black Hole",
			Group:   "Muse",
			Release: release,
			Text:    "Ooh baby, don't you know I suffer?\nOoh baby, can you hear me moan?\nYou caught me under false pretenses\nHow long before you let me go?\n\nOoh\nYou set my soul alight\nOoh\nYou set my soul alight",
			Link:    "https://www.youtube.com/watch?v=Xsp3_a-PMTw",
		}, nil
	}

	log := logger.GetLoggerFromContextOrDefault(ctx).WithGroup(loogerGroup)
	log.Debug("song not found on remote server", "op", op, "song", song)
	return zero, model.ErrNotFound
}
