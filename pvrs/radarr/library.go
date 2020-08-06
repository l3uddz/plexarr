package radarr

import (
	"encoding/json"
	"fmt"
	"github.com/l3uddz/plexarr"
	"net/http"
)

type movieItem struct {
	Title      string  `json:"title"`
	Path       string  `json:"path"`
	ImdbId     *string `json:"imdbId"`
	TmdbId     *uint64 `json:"tmdbId"`
	Downloaded bool    `json:"downloaded"`
	HasFile    bool    `json:"hasFile"`
	Status     string  `json:"status"`
}

func (c *Client) GetLibraryItems() (map[string]plexarr.PvrItem, error) {
	// create request
	req, err := http.NewRequest("GET", plexarr.JoinURL(c.url, "api", "movie"), nil)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", err, plexarr.ErrFatal)
	}

	// set headers
	req.Header.Set("X-Api-Key", c.token)
	req.Header.Set("Accept", "application/json")

	// send request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed retrieving radarr library: %w", err)
	}

	defer res.Body.Close()

	// validate response
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("failed validating radarr library response: %v", res.StatusCode)
	}

	// decode response
	radarrItems := make([]movieItem, 0)
	if err := json.NewDecoder(res.Body).Decode(&radarrItems); err != nil {
		return nil, fmt.Errorf("failed decoding radarr library response: %w", err)
	}

	// create response
	skipNonUniqueItems := make(map[string]int)
	pvrItems := make(map[string]plexarr.PvrItem)
	for _, item := range radarrItems {
		// skip item if we do not have a file or its marked as deleted
		if !item.HasFile || item.Status == "deleted" {
			continue
		}

		// create guids
		guids := make([]string, 0)

		if item.ImdbId != nil && *item.ImdbId != "" {
			guids = append(guids, fmt.Sprintf("com.plexapp.agents.imdb://%s", *item.ImdbId))
		}

		if item.TmdbId != nil && *item.TmdbId != 0 {
			guids = append(guids, fmt.Sprintf("com.plexapp.agents.themoviedb://%d", *item.TmdbId))
		}

		if len(guids) == 0 {
			c.log.Warn().
				Interface("movie", item).
				Msg("Failed creating at-least one plex guid, skipping item")
			continue
		}

		// rewrite path
		rewritePath := c.rewrite(item.Path)

		// skip this item?
		if skips, ok := skipNonUniqueItems[rewritePath]; ok {
			// this item should be skipped
			c.log.Warn().
				Interface("movie", item).
				Str("rewrite_path", rewritePath).
				Str("pvr_path", item.Path).
				Int("path_duplicates", skips).
				Msg("Path is not unique, skipping item(s)")

			skipNonUniqueItems[rewritePath]++
			continue
		}

		// item path exists in map?
		if _, ok := pvrItems[rewritePath]; ok {
			c.log.Warn().
				Interface("movie", item).
				Msg("Path is not unique, skipping item(s)")

			skipNonUniqueItems[rewritePath] = 2
			delete(pvrItems, rewritePath)
			continue
		}

		// add item
		pvrItems[rewritePath] = plexarr.PvrItem{
			Title:   item.Title,
			Path:    rewritePath,
			PvrPath: item.Path,
			GUID:    guids,
		}
	}

	return pvrItems, nil
}
