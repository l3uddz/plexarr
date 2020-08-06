package radarr

import (
	"errors"
	"github.com/l3uddz/plexarr"
	"github.com/rs/zerolog"
)

type Config struct {
	Name   string `yaml:"name"`
	URL    string `yaml:"url"`
	ApiKey string `yaml:"api_key"`

	Verbosity string          `yaml:"verbosity"`
	Rewrite   plexarr.Rewrite `yaml:"rewrite"`
}

type Client struct {
	url   string
	token string

	log     zerolog.Logger
	rewrite plexarr.Rewriter
}

func New(c Config, libraryType plexarr.LibraryType) (*Client, error) {
	if libraryType != plexarr.MovieLibrary {
		return nil, errors.New("only movie libraries are supported")
	}

	rewriter, err := plexarr.NewRewriter(c.Rewrite)
	if err != nil {
		return nil, err
	}

	l := plexarr.GetLogger(c.Verbosity).With().
		Str("pvr", c.Name).
		Str("url", c.URL).Logger()

	return &Client{
		url:     c.URL,
		token:   c.ApiKey,
		log:     l,
		rewrite: rewriter,
	}, nil
}
