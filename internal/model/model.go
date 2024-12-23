package model

// TODO: easyjson

type SongDetail struct {
	ID      uint64 `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Group   string `json:"group,omitempty"`
	Release Date   `json:"release,omitempty"`
	Text    string `json:"text,omitempty"`
	Link    string `json:"link,omitempty"`
}

type GetSongTextRequest struct {
	ID     uint64  `json:"id,omitempty"`
	Offset *uint64 `json:"offset,omitempty"`
	Limit  *uint64 `json:"limit,omitempty"`
}

type SongFilters struct {
	Name    *string `json:"name,omitempty"`
	Group   *string `json:"group,omitempty"`
	Release *Date   `json:"release,omitempty"`
	Text    *string `json:"text,omitempty"`
	Link    *string `json:"link,omitempty"`
	Offset  *uint64 `json:"offset,omitempty"`
	Limit   *uint64 `json:"limit,omitempty"`
}

type SongUpdate struct {
	ID      uint64  `json:"id,omitempty"`
	Name    *string `json:"name,omitempty"`
	Group   *string `json:"group,omitempty"`
	Release *Date   `json:"release,omitempty"`
	Text    *string `json:"text,omitempty"`
	Link    *string `json:"link,omitempty"`
}
