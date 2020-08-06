package plexarr

import (
	"github.com/pkg/errors"
	"regexp"
)

type Pvr interface {
	GetLibraryItems() (map[string]PvrItem, error)
}

type PvrItem struct {
	Title   string
	Path    string
	PvrPath string
	GUID    []string
}

type LibraryType int

var (
	MovieLibrary LibraryType = 1
	TvLibrary    LibraryType = 2
)

var (
	// ErrPlexUnavailable may occur when a plex api cannot be validated
	ErrPlexUnavailable = errors.New("plex unavailable")

	// ErrFatal indicates a severe problem related to development.
	ErrFatal = errors.New("fatal development related error")
)

type Rewrite struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

type Rewriter func(string) string

func NewRewriter(r Rewrite) (Rewriter, error) {
	if r.From == "" || r.To == "" {
		rewriter := func(input string) string {
			return input
		}

		return rewriter, nil
	}

	re, err := regexp.Compile(r.From)
	if err != nil {
		return nil, err
	}

	rewriter := func(input string) string {
		return re.ReplaceAllString(input, r.To)
	}

	return rewriter, nil
}
