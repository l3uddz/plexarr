package main

import (
	"errors"
	"fmt"
	"github.com/l3uddz/plexarr"
	"github.com/l3uddz/plexarr/pvrs/radarr"
	"github.com/l3uddz/plexarr/pvrs/sonarr"
	"strings"
)

func getPvr(name string, cfg config, libraries []plexLibraryItem) (plexarr.Pvr, error) {
	// radarr
	for _, pvr := range cfg.Pvr.Radarr {
		if !strings.EqualFold(name, pvr.Name) {
			continue
		}

		// validate all libraries are movies
		for _, lib := range libraries {
			if lib.Type != plexarr.MovieLibrary {
				return nil, errors.New("radarr only supports movie libraries")
			}
		}

		// init pvr object
		p, err := radarr.New(pvr)
		if err != nil {
			return nil, fmt.Errorf("failed initialising radarr pvr %v: %w", pvr.Name, err)
		}

		return p, nil
	}

	// sonarr
	for _, pvr := range cfg.Pvr.Sonarr {
		if !strings.EqualFold(name, pvr.Name) {
			continue
		}

		// validate all libraries are series
		for _, lib := range libraries {
			if lib.Type != plexarr.MovieLibrary {
				return nil, errors.New("sonarr only supports tv libraries")
			}
		}

		// init pvr object
		p, err := sonarr.New(pvr)
		if err != nil {
			return nil, fmt.Errorf("failed initialising sonarr pvr %v: %w", pvr.Name, err)
		}

		return p, nil
	}

	return nil, errors.New("pvr not found")
}
