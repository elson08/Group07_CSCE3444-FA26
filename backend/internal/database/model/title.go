package model

const (
	KindMovie string = "movie"
	KindShow  string = "show"
)

// Title is a movie or show in the catalog.
type Title struct {
	ID            int64     `json:"id"`
	ProviderID    *string   `json:"provider_id,omitempty"`
	Kind          string    `json:"kind"`
	Name          string    `json:"name"`
	ReleaseYear   *int      `json:"release_year,omitempty"`
	Synopsis      *string   `json:"synopsis,omitempty"`
	PosterURL     *string   `json:"poster_url,omitempty"`
	TrailerURL    *string   `json:"trailer_url,omitempty"`
	RatingAverage *float64  `json:"rating_average,omitempty"`
	RatingCount   int       `json:"rating_count"`
	Genres        []Genre   `json:"genres,omitempty"`
	Services      []Service `json:"services,omitempty"`
}

type TitleDetails struct {
	Synopsis   string
	PosterURL  string
	TrailerURL string
	Services   []string
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
