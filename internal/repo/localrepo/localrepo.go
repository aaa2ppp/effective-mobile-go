package localrepo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"effective-mobile-go/internal/model"
)

var (
	ErrNotFound      = model.ErrNotFound
	ErrInternalError = model.ErrInternalError
)

type LocalRepo struct {
	db *sql.DB
}

func New(db *sql.DB) LocalRepo {
	return LocalRepo{
		db: db,
	}
}

type SongDetail = model.SongDetail

// CreateSong всегда возвращает детальную информацию о песне. Если в базе нет группы или песни,
// то они будут созданы на основаннии входящих данных.
func (r LocalRepo) CreateSong(ctx context.Context, song SongDetail) (SongDetail, error) {
	var zero SongDetail
	x := newHelper(ctx, "CreateSong")

	const q = `
		WITH
		ins_group AS (
			INSERT INTO "group" (name) VALUES ($2)
			ON CONFLICT(name) DO NOTHING
			RETURNING id, name
		)
		,ins_or_sel_group AS (
			SELECT id, name FROM ins_group
			UNION
			SELECT id, name FROM "group" WHERE name = $2
		)
		,ins_song(id, name, group_id, release, text, link) AS (
			INSERT INTO song (name, group_id, release, text, link)
			SELECT $1, (SELECT id FROM ins_or_sel_group), $3, $4, $5
			ON CONFLICT(name, group_id) DO NOTHING
			RETURNING id, name, group_id, release, text, link
		)
		,ins_or_sel_song AS (
			SELECT id, name, group_id, release, link FROM ins_song 
			UNION
			SELECT id, name, group_id, release, link FROM song
			WHERE name = $1 AND group_id = (SELECT id FROM ins_or_sel_group)
		)
		SELECT s.id, s.name, g.name, s.release, s.link
		FROM ins_or_sel_song AS s, ins_or_sel_group AS g
	`

	err := r.db.QueryRowContext(ctx, q, song.Name, song.Group, song.Release.Time, song.Text, song.Link).
		Scan(&song.ID, &song.Name, &song.Group, &song.Release.Time, &song.Link)

	if err != nil {
		x.Log().Error("can't query", "error", err, "query", q, "song", song)
		return zero, ErrInternalError
	}

	return song, nil
}

// GetSong возвращает детальную информацию о песне по ID. Есле в базе нет такого ID возвращает ErrNotFound.
func (r LocalRepo) GetSong(ctx context.Context, songID uint64) (SongDetail, error) {
	x := newHelper(ctx, "GetSong")
	var zero SongDetail

	const q = `
		SELECT s.id, s.name, g.name, s.release, s.text, s.link
		FROM song AS s JOIN "group" AS g ON s.group_id = g.id
		WHERE s.id = $1;
	`

	var song SongDetail

	err := r.db.QueryRowContext(ctx, q, songID).
		Scan(&song.ID, &song.Name, &song.Group, &song.Release.Time, &song.Text, &song.Link)

	if err != nil {

		if err == sql.ErrNoRows {
			return zero, ErrNotFound
		}

		x.Log().Error("can't query", "error", err, "query", q, "songID", songID)
		return zero, ErrInternalError
	}

	return song, nil
}

type SongListFilters = model.SongFilters

func (r LocalRepo) ListSongs(ctx context.Context, req SongListFilters) ([]SongDetail, error) {
	x := newHelper(ctx, "ListSongs")
	var zero []SongDetail

	var q = `
		SELECT s.id, s.name, g.name, s.release, link
		FROM song AS s JOIN "group" AS g ON s.group_id = g.id
		%s /* where placeholder */
		ORDER BY s.id
		%s /* limit placeholder */
		%s /* offcet placeholder */
	`
	var (
		idx     int
		filters []string
		values  []any
	)

	if req.Name != nil {
		idx++
		filters = append(filters, fmt.Sprintf(`s.name = $%d`, idx))
		values = append(values, *req.Name)
	}
	if req.Group != nil {
		idx++
		filters = append(filters, fmt.Sprintf(`g.name = $%d`, idx))
		values = append(values, req.Group)
	}
	if req.Text != nil {
		idx++
		filters = append(filters, fmt.Sprintf(`s.text = %%$%d%%`, idx))
		values = append(values, req.Text)
	}
	if req.Release != nil {
		idx++
		filters = append(filters, fmt.Sprintf(`s.release = $%d`, idx))
		values = append(values, req.Release)
	}
	if req.Link != nil {
		idx++
		filters = append(filters, fmt.Sprintf(`s.link = $%d`, idx))
		values = append(values, req.Link)
	}

	q = fmt.Sprintf(q,
		func() string {
			if len(filters) != 0 {
				return "WHERE " + strings.Join(filters, " AND ")
			}
			return ""
		}(),
		func() string {
			if req.Limit != nil {
				idx++
				values = append(values, req.Limit)
				return fmt.Sprintf("LIMIT $%d", idx)
			}
			return ""
		}(),
		func() string {
			if req.Offset != nil {
				idx++
				values = append(values, req.Offset)
				return fmt.Sprintf("LIMIT $%d", idx)
			}
			return ""
		}(),
	)

	rows, err := r.db.QueryContext(ctx, q, values...)
	if err != nil {

		x.Log().Error("can't query", "error", err, "query", q, "req", req)
		return zero, ErrInternalError
	}

	defer rows.Close()

	var (
		song SongDetail
		resp []SongDetail
	)

	for rows.Next() {
		if err := rows.Scan(&song.ID, &song.Name, &song.Group, &song.Release.Time, &song.Link); err != nil {

			x.Log().Error("can't scan", "error", err, "query", q)
			return zero, ErrInternalError
		}

		resp = append(resp, song)
	}

	return resp, nil
}

type UpdateSongRequest = model.SongUpdate

// UpdateSong обновляет информацию о песне с указанным ID. Возвращает детальную информацию о песне.
// Если песня с указанным ID отсутствует в базе, то возвращает ErrNotFound.
func (r LocalRepo) UpdateSong(ctx context.Context, req UpdateSongRequest) (SongDetail, error) {
	x := newHelper(ctx, "UpdateSong")

	var q = `
		UPDATE song SET (%s /* fields */) = (%s /* placeholders */) id = $1
		RETURNING s.id, s.name, g.name, s.release, s.text, s.link;
	`

	var (
		idx         int
		fields      []string
		paceholders []string
		values      []any
	)

	idx++
	values = append(values, req.ID)

	if req.Name != nil {
		idx++
		fields = append(fields, "name")
		values = append(values, req.Name)
		paceholders = append(paceholders, fmt.Sprintf("$%d", idx))
	}
	if req.Release != nil {
		idx++
		fields = append(fields, "release")
		values = append(values, req.Release)
		paceholders = append(paceholders, fmt.Sprintf("$%d", idx))
	}
	if req.Text != nil {
		idx++
		fields = append(fields, "text")
		values = append(values, req.Text)
		paceholders = append(paceholders, fmt.Sprintf("$%d", idx))
	}
	if req.Link != nil {
		idx++
		fields = append(fields, "link")
		values = append(values, req.Text)
		paceholders = append(paceholders, fmt.Sprintf("$%d", idx))
	}

	if len(fields) == 0 {
		return r.GetSong(ctx, req.ID)
	}

	var zero SongDetail
	q = fmt.Sprintf(q, strings.Join(fields, ","), strings.Join(paceholders, ","))

	var song SongDetail

	err := r.db.QueryRowContext(ctx, q, values...).
		Scan(&song.ID, &song.Name, &song.Group, &song.Release.Time, &song.Text, &song.Link)

	if err != nil {

		if err == sql.ErrNoRows {
			return zero, ErrNotFound
		}

		x.Log().Error("can't query", "error", err, "query", q, "req", req)
		return zero, ErrInternalError
	}

	return song, nil
}

// DeleteSong удадаляет песню с указанным ID, если она есть в базе. НЕ возвращает ошибку,
// если песни нет в базе. Иными словами, гарантируется, что в случае успешного завершения,
// в базе нет песни с указанным ID.
func (r LocalRepo) DeleteSong(ctx context.Context, songID uint64) error {
	x := newHelper(ctx, "DeleteSong")

	const q = `DELETE FROM song WHERE id = $1`

	if _, err := r.db.ExecContext(ctx, q, songID); err != nil {

		x.Log().Error("can't query", "error", err, "query", q, "songID", songID)
		return ErrInternalError
	}

	return nil
}
