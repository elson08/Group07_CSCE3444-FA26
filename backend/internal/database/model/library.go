package model

type TitleFilter struct {
	Query          string
	Kind           string
	Genres         []string
	ServiceIDs     []int64
	ServicesOfUser int64
	YearFrom       int
	YearTo         int
	MinRating      float64
	Sort           string
	Limit          int
	Offset         int
}

type Rating struct {
	Title     Title  `json:"title"`
	Stars     int    `json:"stars"`
	Review    string `json:"review"`
	UpdatedAt string `json:"updated_at"`
}

type Review struct {
	Username  string `json:"username"`
	Stars     int    `json:"stars"`
	Review    string `json:"review"`
	UpdatedAt string `json:"updated_at"`
}

type TitleStatus struct {
	Saved   bool   `json:"saved"`
	Watched bool   `json:"watched"`
	Stars   int    `json:"stars"`
	Review  string `json:"review"`
}
