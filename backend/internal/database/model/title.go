package model

const (
	KindMovie string = "movie"
	KindShow  string = "show"
)

// Title is a movie or show in the catalog.
type Title struct {
	ID          int64     `json:"id"`
	ProviderID  *string   `json:"provider_id,omitempty"`
	Kind        string    `json:"kind"`
	Name        string    `json:"name"`
	ReleaseYear *int      `json:"release_year,omitempty"`
	Synopsis    *string   `json:"synopsis,omitempty"`
	PosterURL   *string   `json:"poster_url,omitempty"`
	TrailerURL  *string   `json:"trailer_url,omitempty"`
	Genres      []Genre   `json:"genres,omitempty"`
	Services    []Service `json:"services,omitempty"`
}

// Genre categorizes titles.
type Genre struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Service identifies a streaming provider.
type Service struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
