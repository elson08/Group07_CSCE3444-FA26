package database

import (
	"bufio"
	"compress/gzip"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/elson08/Group07_CSCE3444-FA26/backend/internal/database/model"
)

const (
	imdbBaseURL  = "https://datasets.imdbws.com/"
	imdbMinVotes = 1000
	imdbTimeout  = 15 * time.Minute
	imdbNull     = `\N`

	imdbRatingsColumns = 3
	imdbBasicsColumns  = 9
)

var imdbKinds = map[string]string{
	"movie":        model.KindMovie,
	"tvSeries":     model.KindShow,
	"tvMiniSeries": model.KindShow,
}

func (d *Database) SeedCatalog(ctx context.Context) (int, error) {
	return d.seedCatalog(ctx, imdbBaseURL)
}

func (d *Database) seedCatalog(ctx context.Context, baseURL string) (int, error) {
	var hasTitles bool
	if err := d.conn.QueryRowContext(ctx, hasTitlesQuery).Scan(&hasTitles); err != nil {
		return 0, err
	}

	if hasTitles {
		return 0, nil
	}

	log.Print("catalog is empty, importing from IMDb (this can take a few minutes)")

	ctx, cancel := context.WithTimeout(ctx, imdbTimeout)
	defer cancel()

	popular, err := popularTitles(ctx, baseURL+"title.ratings.tsv.gz")
	if err != nil {
		return 0, err
	}

	tx, err := d.conn.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	count, err := importTitles(ctx, tx, baseURL+"title.basics.tsv.gz", popular)
	if err != nil {
		return 0, err
	}

	if count == 0 {
		return 0, errors.New("IMDb data contained no matching titles")
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return count, nil
}

func popularTitles(ctx context.Context, url string) (map[string]struct{}, error) {
	popular := map[string]struct{}{}

	err := readIMDbFile(ctx, url, imdbRatingsColumns, func(row []string) error {
		if votes, err := strconv.Atoi(row[2]); err == nil && votes >= imdbMinVotes {
			popular[row[0]] = struct{}{}
		}

		return nil
	})

	return popular, err
}

func importTitles(ctx context.Context, tx *sql.Tx, url string, popular map[string]struct{}) (int, error) {
	insertTitle, err := tx.PrepareContext(ctx, insertTitleQuery)
	if err != nil {
		return 0, err
	}
	defer insertTitle.Close()

	upsertGenre, err := tx.PrepareContext(ctx, upsertGenreQuery)
	if err != nil {
		return 0, err
	}
	defer upsertGenre.Close()

	linkGenre, err := tx.PrepareContext(ctx, linkTitleGenreQuery)
	if err != nil {
		return 0, err
	}
	defer linkGenre.Close()

	genreIDs := map[string]int64{}
	count := 0

	err = readIMDbFile(ctx, url, imdbBasicsColumns, func(row []string) error {
		providerID, titleType, name, isAdult, startYear, genres := row[0], row[1], row[2], row[4], row[5], row[8]

		kind, wanted := imdbKinds[titleType]
		_, isPopular := popular[providerID]
		if !wanted || !isPopular || isAdult == "1" {
			return nil
		}

		var year *int
		if parsed, err := strconv.Atoi(startYear); err == nil {
			year = &parsed
		}

		result, err := insertTitle.ExecContext(ctx, providerID, kind, name, year)
		if err != nil {
			return err
		}

		titleID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		count++

		if genres == imdbNull {
			return nil
		}

		for _, genre := range strings.Split(genres, ",") {
			genreID, known := genreIDs[genre]
			if !known {
				if err := upsertGenre.QueryRowContext(ctx, genre).Scan(&genreID); err != nil {
					return err
				}
				genreIDs[genre] = genreID
			}

			if _, err := linkGenre.ExecContext(ctx, titleID, genreID); err != nil {
				return err
			}
		}

		return nil
	})

	return count, err
}

func readIMDbFile(ctx context.Context, url string, columns int, handle func([]string) error) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: %s", url, response.Status)
	}

	archive, err := gzip.NewReader(response.Body)
	if err != nil {
		return err
	}
	defer archive.Close()

	scanner := bufio.NewScanner(archive)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	scanner.Scan()
	for scanner.Scan() {
		row := strings.Split(scanner.Text(), "\t")
		if len(row) != columns {
			continue
		}

		if err := handle(row); err != nil {
			return err
		}
	}

	return scanner.Err()
}
