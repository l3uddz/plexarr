package plex

import (
	"fmt"
	"github.com/l3uddz/plexarr"
	"net/http"
	"net/url"
	"strconv"
)

func (c *Client) Available() error {
	// create request
	req, err := http.NewRequest("GET", plexarr.JoinURL(c.url, "myplex", "account"), nil)
	if err != nil {
		return fmt.Errorf("%v: %w", err, plexarr.ErrFatal)
	}

	// set headers
	req.Header.Set("X-Plex-Token", c.token)
	req.Header.Set("Accept", "application/json")

	// send request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not check Plex availability: %v: %w",
			err, plexarr.ErrPlexUnavailable)
	}

	defer res.Body.Close()

	// validate response
	if res.StatusCode != 200 {
		return fmt.Errorf("could not check Plex availability: %v: %w",
			res.StatusCode, plexarr.ErrPlexUnavailable)
	}

	return nil
}

func (c *Client) Split(metadataItemId int) error {
	// create request
	req, err := http.NewRequest("PUT",
		plexarr.JoinURL(c.url, "library", "metadata", strconv.Itoa(metadataItemId), "split"), nil)
	if err != nil {
		return fmt.Errorf("%v: %w", err, plexarr.ErrFatal)
	}

	// set headers
	req.Header.Set("X-Plex-Token", c.token)

	// send request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not split Plex metadata_item %v: %v: %w",
			metadataItemId, err, plexarr.ErrPlexUnavailable)
	}

	defer res.Body.Close()

	// validate response
	if res.StatusCode != 200 {
		return fmt.Errorf("could not split Plex metadata_item %v: %v: %w",
			metadataItemId, res.StatusCode, plexarr.ErrFatal)
	}

	return nil
}

func (c *Client) Match(metadataItemId int, title string, guid string) error {
	// create request
	req, err := http.NewRequest("PUT",
		plexarr.JoinURL(c.url, "library", "metadata", strconv.Itoa(metadataItemId), "match"), nil)
	if err != nil {
		return fmt.Errorf("%v: %w", err, plexarr.ErrFatal)
	}

	// set headers
	req.Header.Set("X-Plex-Token", c.token)

	// set params
	q := url.Values{}
	q.Set("guid", guid)
	q.Set("name", title)

	req.URL.RawQuery = q.Encode()

	// send request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not match Plex metadata_item %v: %v: %w",
			metadataItemId, err, plexarr.ErrPlexUnavailable)
	}

	defer res.Body.Close()

	// validate response
	if res.StatusCode != 200 {
		return fmt.Errorf("could not match Plex metadata_item %v: %v: %w",
			metadataItemId, res.StatusCode, plexarr.ErrFatal)
	}

	return nil
}
