package plex

import (
	"github.com/l3uddz/plexarr"
	"github.com/rs/zerolog"
)

type Config struct {
	URL      string          `yaml:"url"`
	Token    string          `yaml:"token"`
	Database string          `yaml:"database"`
	Rewrite  plexarr.Rewrite `yaml:"rewrite"`

	Verbosity string `yaml:"verbosity"`
}

type Client struct {
	url       string
	token     string
	libraries []library

	log   zerolog.Logger
	store *datastore
}

func New(c Config) (*Client, error) {

	store, err := newDatastore(c.Database)
	if err != nil {
		return nil, err
	}

	libraries, err := store.Libraries()
	if err != nil {
		return nil, err
	}

	l := plexarr.GetLogger(c.Verbosity).With().
		Str("url", c.URL).Logger()

	l.Debug().
		Interface("libraries", libraries).
		Msg("Retrieved libraries")

	return &Client{
		url:       c.URL,
		token:     c.Token,
		libraries: libraries,

		log:   l,
		store: store,
	}, nil
}
