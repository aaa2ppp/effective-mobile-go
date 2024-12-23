package remoterepo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"effective-mobile-go/internal/config"
	"effective-mobile-go/internal/lib"
	"effective-mobile-go/internal/logger"
	"effective-mobile-go/internal/model"
)

type SongDetail = model.SongDetail

var (
	ErrNotFound      = model.ErrNotFound
	ErrBadRequest    = model.ErrBadRequest
	ErrInternalError = model.ErrInternalError
)

const loggerGroup = "remoterepo"

type RemoteRepo struct {
	url string
}

func New(cfg config.RemoteAPIConfig) RemoteRepo {
	return RemoteRepo{
		url: cfg.URL,
	}
}

func (r RemoteRepo) GetSong(ctx context.Context, song SongDetail) (SongDetail, error) {
	const op = "GetSong"
	var zero SongDetail

	client := http.Client{Timeout: 10 * time.Second} // TODO: timeout to config

	url := fmt.Sprintf("%s?group=%s&song=%s", r.url, url.QueryEscape(song.Group),
		url.QueryEscape(song.Name))

	resp, err := client.Get(url)
	if err != nil {

		log(ctx).Error("can't get ulr", "error", err, "url", url)
		return zero, ErrInternalError
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {

		var err error
		switch resp.StatusCode {
		case http.StatusBadRequest:
			err = ErrBadRequest
		case http.StatusNotFound:
			err = ErrNotFound
		default:
			err = ErrInternalError
		}

		body, _ := io.ReadAll(resp.Body)
		log(ctx).Error("remote server returned not OK status", "op", op, "url", url,
			"statusCode", resp.StatusCode, "body", lib.UnsafeString(body))
		return zero, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {

		log(ctx).Error("can't read response body", "op", op, "error", err, "url", url)
		return zero, ErrInternalError
	}

	if err := json.Unmarshal(body, &song); err != nil {

		log(ctx).Error("can't parse response body", "op", op, "error", err, "url", url,
			"body", lib.UnsafeString(body))
		return zero, ErrInternalError
	}

	song.ID = 0 // for security

	return song, nil
}

func log(ctx context.Context) *slog.Logger {
	return logger.GetLoggerFromContextOrDefault(ctx).WithGroup(loggerGroup)
}
