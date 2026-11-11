package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/elson08/Group07_CSCE3444-FA26/backend/internal/database/model"
)

const (
	tmdbBaseURL    = "https://api.themoviedb.org/3"
	tmdbPosterURL  = "https://image.tmdb.org/t/p/w500"
	tmdbTrailerURL = "https://www.youtube.com/embed/"
	tmdbRegion     = "US"
	tmdbParallel   = 8
)

type tmdbClient struct {
	token   string
	baseURL string
	client  *http.Client
}

type tmdbMatch struct {
	ID         int64  `json:"id"`
	Overview   string `json:"overview"`
	PosterPath string `json:"poster_path"`
}

func newTMDBClient(token string) *tmdbClient {
	if token == "" {
		log.Print("tmdb_token is not set, titles will have no posters or descriptions")
		return nil
	}

	return &tmdbClient{
		token:   token,
		baseURL: tmdbBaseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *tmdbClient) get(ctx context.Context, path string, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+c.token)

	response, err := c.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", path, response.Status)
	}

	return json.NewDecoder(response.Body).Decode(target)
}

func (c *tmdbClient) lookup(ctx context.Context, imdbID, kind string) (model.TitleDetails, error) {
	var details model.TitleDetails
	var found struct {
		Movies []tmdbMatch `json:"movie_results"`
		Shows  []tmdbMatch `json:"tv_results"`
	}

	if err := c.get(ctx, "/find/"+url.PathEscape(imdbID)+"?external_source=imdb_id", &found); err != nil {
		return details, err
	}

	matches, mediaType := found.Movies, "movie"
	if kind == model.KindShow {
		matches, mediaType = found.Shows, "tv"
	}

	if len(matches) == 0 {
		return details, nil
	}

	match := matches[0]
	details.Synopsis = match.Overview
	if match.PosterPath != "" {
		details.PosterURL = tmdbPosterURL + match.PosterPath
	}

	var videos struct {
		Results []struct {
			Key  string `json:"key"`
			Site string `json:"site"`
			Type string `json:"type"`
		} `json:"results"`
	}

	mediaPath := fmt.Sprintf("/%s/%d", mediaType, match.ID)
	if err := c.get(ctx, mediaPath+"/videos", &videos); err != nil {
		return details, err
	}

	for _, video := range videos.Results {
		if video.Site == "YouTube" && video.Type == "Trailer" {
			details.TrailerURL = tmdbTrailerURL + video.Key
			break
		}
	}

	var providers struct {
		Results map[string]struct {
			Flatrate []struct {
				Name string `json:"provider_name"`
			} `json:"flatrate"`
		} `json:"results"`
	}

	if err := c.get(ctx, mediaPath+"/watch/providers", &providers); err != nil {
		return details, err
	}

	for _, provider := range providers.Results[tmdbRegion].Flatrate {
		details.Services = append(details.Services, provider.Name)
	}

	return details, nil
}

func (s *Server) fillDetails(ctx context.Context, titles []model.Title) bool {
	if s.tmdb == nil {
		return false
	}

	var waiting sync.WaitGroup
	var filled atomic.Bool
	slots := make(chan struct{}, tmdbParallel)

	for _, title := range titles {
		if title.Synopsis != nil || title.ProviderID == nil {
			continue
		}

		waiting.Add(1)
		slots <- struct{}{}

		go func() {
			defer waiting.Done()
			defer func() { <-slots }()

			details, err := s.tmdb.lookup(ctx, *title.ProviderID, title.Kind)
			if err == nil {
				err = s.database.SaveTitleDetails(ctx, title.ID, details)
			}

			if err != nil {
				log.Printf("TMDB lookup failed for %s: %v", *title.ProviderID, err)
				return
			}

			filled.Store(true)
		}()
	}

	waiting.Wait()
	return filled.Load()
}
