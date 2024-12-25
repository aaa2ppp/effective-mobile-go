package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"effective-mobile-go/internal/model"
)

var (
	ErrBadRequest    = model.ErrBadRequest
	ErrNotFound      = model.ErrNotFound
	ErrInternalError = model.ErrInternalError
)

type Service interface {
	CreateSong(context.Context, model.SongDetail) (model.SongDetail, error)
	ListSongs(context.Context, model.SongFilters) ([]model.SongDetail, error)
	GetSong(_ context.Context, songID uint64) (model.SongDetail, error)
	GetSongText(context.Context, model.GetSongTextRequest) ([]string, error)
	UpdateSong(context.Context, model.SongUpdate) (model.SongDetail, error)
	DeleteSong(_ context.Context, songID uint64) error
}

func New(service Service) http.Handler {

	mux := http.NewServeMux()

	mux.Handle("GET /ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		x := newHelper("ping", w, r)
		x.Log().Debug("knock-knock", slog.String("remoteAddr", r.RemoteAddr))
		x.WriteResponse("pong")
	}))

	h := handler{service}

	mux.Handle("GET    /songs", http.HandlerFunc(h.listSongsHandler))
	mux.Handle("POST   /songs", http.HandlerFunc(h.createSongHandler))

	mux.Handle("GET    /songs/{id}", http.HandlerFunc(h.getSongHandler))
	mux.Handle("GET    /songs/{id}/text", http.HandlerFunc(h.getSongTextHandler))
	mux.Handle("POST   /songs/{id}", http.HandlerFunc(h.updateSongHandler))
	mux.Handle("DELETE /songs/{id}", http.HandlerFunc(h.deleteSongHandler))

	return mux
}

type handler struct {
	Service
}

type emptyResponse struct{}

type songDetail struct {
	ID      uint64 `json:"id"`
	Name    string `json:"name,omitempty"`
	Group   string `json:"group,omitempty"`
	Release string `json:"release,omitempty" example:"02.01.2006"`
	Text    string `json:"text,omitempty"`
	Link    string `json:"link,omitempty"`
}

type createSongRequest struct {
	Song  string `json:"song" example:"Supermassive Black Hole"`
	Group string `json:"group" example:"Muse"`
}

type createSongResponse struct {
	Song songDetail `json:"song,omitempty"`
}

// createSongHandler godoc
//
//	@Summary	Create song enrty
//	@Tags		songs
//	@Accept		json
//	@Produce	json
//	@Param		req	body		createSongRequest	true	"CreateSongRequest"
//	@Success	200	{object}	createSongResponse
//	@Failure	400	{object}	errorResponse
//	@Failure	404	{object}	errorResponse
//	@Failure	500	{object}	errorResponse
//	@Router		/songs [post]
func (h handler) createSongHandler(w http.ResponseWriter, r *http.Request) {
	x := newHelper("createSongHandler", w, r)

	var req createSongRequest

	if err := x.DecodeBody(&req); err != nil {
		x.WriteError(err)
		return
	}

	if req.Song == "" {

		x.Log().Debug("song is required")
		x.WriteError(ErrBadRequest)
		return
	}

	if req.Group == "" {

		x.Log().Debug("group is required")
		x.WriteError(ErrBadRequest)
		return
	}

	song := model.SongDetail{
		Name:  req.Song,
		Group: req.Group,
	}

	x.Log().Debug("http request parsed", "song", song)

	song, err := h.CreateSong(x.Ctx(), song)
	if err != nil {
		x.WriteError(err)
		return
	}

	resp := createSongResponse{
		Song: songDetail{
			ID:      song.ID,
			Name:    song.Name,
			Group:   song.Group,
			Release: song.Release.String(),
			Link:    song.Link,
		},
	}

	x.WriteResponse(&resp)
}

type listSongsResponse struct {
	Songs []songDetail
}

// listSongsHandler godoc
//
//	@Summary	List song library
//	@Tags		songs
//	@Produce	json
//	@Param		song	query		string	false	"Song name"
//	@Param		group	query		string	false	"Song group name"
//	@Param		reliase	query		string	false	"Song release date (example: 02.01.2006)"
//	@Param		text	query		string	false	"Song text should contain it"
//	@Param		link	query		string	false	"Song link"
//	@Param		offset	query		uint64	false	"Offeset"
//	@Param		limit	query		uint64	false	"Limit"
//	@Success	200		{object}	listSongsResponse
//	@Failure	400		{object}	errorResponse
//	@Failure	404		{object}	errorResponse
//	@Failure	500		{object}	errorResponse
//	@Router		/songs [get]
func (h handler) listSongsHandler(w http.ResponseWriter, r *http.Request) {
	x := newHelper("listSongsHandler", w, r)

	var req model.SongFilters

	{
		s := r.FormValue("song")
		if s != "" {
			req.Name = &s
		}
	}
	{
		s := r.FormValue("group")
		if s != "" {
			req.Group = &s
		}
	}
	{
		s := r.FormValue("release")
		if s != "" {
			v, err := model.ParseDate(s)
			if err != nil {

				x.Log().Debug("can't parse release date", "error", err, "release", s)
				x.WriteError(ErrBadRequest)
				return
			}
			req.Release = &v
		}
	}
	{
		s := r.FormValue("text")
		if s != "" {
			req.Text = &s
		}
	}
	{
		s := r.FormValue("link")
		if s != "" {
			req.Link = &s
		}
	}
	{
		s := r.FormValue("offset")
		if s != "" {
			v, err := strconv.ParseUint(s, 10, 64)
			if err != nil {

				x.Log().Debug("can't parse offset", "error", err)
				x.WriteError(ErrBadRequest)
				return
			}
			req.Offset = &v
		}
	}
	{
		s := r.FormValue("limit")
		if s != "" {
			v, err := strconv.ParseUint(s, 10, 64)
			if err != nil {

				x.Log().Debug("can't parse limit", "error", err)
				x.WriteError(ErrBadRequest)
				return
			}
			if v == 0 {

				x.Log().Debug("limit can not be 0", "error", err)
				x.WriteError(ErrBadRequest)
				return
			}
			req.Limit = &v
		}
	}

	x.Log().Debug("http request parsed", "req", req)

	songs, err := h.ListSongs(x.Ctx(), req)
	if err != nil {
		x.WriteError(err)
		return
	}

	resp := listSongsResponse{Songs: []songDetail{}} // guarantee not nil

	for i := range songs {
		song := &songs[i]
		resp.Songs = append(resp.Songs, songDetail{
			ID:      song.ID,
			Name:    song.Name,
			Group:   song.Group,
			Release: song.Release.String(),
			Link:    song.Link,
		})
	}

	x.WriteResponse(&resp)
}

type getSongResponse struct {
	Song songDetail `json:"song,omitempty"`
}

// getSongHandler godoc
//
//	@Summary	Get song entry by id
//	@Tags		songs
//	@Produce	json
//	@Param		id	path		uint64	true	"Song id"
//	@Success	200	{object}	getSongResponse
//	@Failure	400	{object}	errorResponse
//	@Failure	404	{object}	errorResponse
//	@Failure	500	{object}	errorResponse
//	@Router		/songs/{id} [get]
func (h handler) getSongHandler(w http.ResponseWriter, r *http.Request) {
	x := newHelper("getSongHandler", w, r)

	songID, err := x.GetID()
	if err != nil {
		x.WriteError(err)
		return
	}

	x.Log().Debug("http request parsed", "songID", songID)

	song, err := h.GetSong(x.Ctx(), songID)
	if err != nil {
		x.WriteError(err)
		return
	}

	resp := getSongResponse{
		Song: songDetail{
			ID:      song.ID,
			Name:    song.Name,
			Group:   song.Group,
			Release: song.Release.String(),
			Link:    song.Link,
		},
	}

	x.WriteResponse(&resp)
}

type getSongTextResponse struct {
	Verses []string `json:"verses,omitempty"`
}

// getSongTextHandler godoc
//
//	@Summary	Get song verses text
//	@Tags		songs
//	@Produce	json
//	@Param		id		path		uint	true	"Song id"
//	@Param		offset	query		uint64	false	"Offeset"
//	@Param		limit	query		uint64	false	"Limit"
//	@Success	200		{object}	getSongResponse
//	@Failure	400		{object}	errorResponse
//	@Failure	404		{object}	errorResponse
//	@Failure	500		{object}	errorResponse
//	@Router		/songs/{id}/text [get]
func (h handler) getSongTextHandler(w http.ResponseWriter, r *http.Request) {
	x := newHelper("getSongHandler", w, r)

	x.Log().Debug("getSongTextHandler")

	var req model.GetSongTextRequest

	{
		v, err := x.GetID()
		if err != nil {
			x.WriteError(err)
			return
		}
		req.ID = v
	}
	{
		s := r.FormValue("offset")
		if s != "" {
			v, err := strconv.ParseUint(s, 10, 64)
			if err != nil {

				x.Log().Debug("can't parse offset", "error", err)
				x.WriteError(ErrBadRequest)
				return
			}
			req.Offset = &v
		}
	}
	{
		s := r.FormValue("limit")
		if s != "" {
			v, err := strconv.ParseUint(s, 10, 64)
			if err != nil {

				x.Log().Debug("can't parse limit", "error", err)
				x.WriteError(ErrBadRequest)
				return
			}
			if v == 0 {

				x.Log().Debug("limit can not be 0")
				x.WriteError(ErrBadRequest)
				return
			}
			req.Limit = &v
		}
	}

	x.Log().Debug("http request parsed", "req", req)

	verses, err := h.GetSongText(x.Ctx(), req)
	if err != nil {
		x.WriteError(err)
		return
	}

	resp := getSongTextResponse{
		Verses: verses,
	}

	x.WriteResponse(&resp)
}

type updateSongRequest struct {
	Release *model.Date `json:"release,omitempty" swaggertype:"string" example:"02.01.2006"`
	Text    *string     `json:"text,omitempty"`
	Link    *string     `json:"link,omitempty"`
}

type updateSongResponse struct {
	Song songDetail
}

// updateSongHandler godoc
//
//	@Summary	List song library
//	@Tags		songs
//	@Produce	json
//	@Param		id	path		uint				true	"Song id"
//	@Param		req	body		updateSongRequest	true	"UpdateSongRequest"
//	@Success	200	{object}	updateSongResponse
//	@Failure	400	{object}	errorResponse
//	@Failure	404	{object}	errorResponse
//	@Failure	500	{object}	errorResponse
//	@Router		/songs/{id} [post]
func (h handler) updateSongHandler(w http.ResponseWriter, r *http.Request) {
	x := newHelper("updateSongHandler", w, r)

	songID, err := x.GetID()
	if err != nil {
		x.WriteError(err)
		return
	}

	var req updateSongRequest

	if err := x.DecodeBody(&req); err != nil {
		x.WriteError(err)
		return
	}

	update := model.SongUpdate{
		ID:      songID,
		Release: req.Release,
		Text:    req.Text,
		Link:    req.Link,
	}

	x.Log().Debug("http request parsed", "update", update)

	song, err := h.UpdateSong(x.Ctx(), update)
	if err != nil {
		x.WriteError(err)
		return
	}

	resp := updateSongResponse{
		Song: songDetail{
			ID:      song.ID,
			Name:    song.Name,
			Group:   song.Group,
			Release: song.Release.String(),
			Link:    song.Link,
		},
	}

	x.WriteResponse(&resp)
}

// deleteSongHandler godoc
//
//	@Summary	Delete song library entry
//	@Tags		songs
//	@Produce	json
//	@Param		id	path		uint	true	"Song id"
//	@Success	200	{object}	emptyResponse
//	@Failure	400	{object}	errorResponse
//	@Failure	404	{object}	errorResponse
//	@Failure	500	{object}	errorResponse
//	@Router		/songs/{id} [delete]
func (h handler) deleteSongHandler(w http.ResponseWriter, r *http.Request) {
	x := newHelper("deleteSongHandler", w, r)

	songID, err := x.GetID()
	if err != nil {
		x.WriteError(err)
		return
	}

	x.Log().Debug("http request parsed", "songID", songID)

	if err := h.DeleteSong(x.Ctx(), songID); err != nil {
		x.WriteError(err)
		return
	}

	x.WriteResponse(&emptyResponse{})
}
