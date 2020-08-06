package main

import (
	"errors"
	"fmt"
	"github.com/l3uddz/plexarr"
	"github.com/l3uddz/plexarr/pvrs/radarr"
	"github.com/l3uddz/plexarr/pvrs/sonarr"
	"strings"
)

func getPvr(name string, cfg config, libraryType plexarr.LibraryType) (plexarr.Pvr, error) {
	// radarr
	for _, pvr := range cfg.Pvr.Radarr {
		if !strings.EqualFold(name, pvr.Name) {
			continue
		}

		// init pvr object
		p, err := radarr.New(pvr, libraryType)
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

		// init pvr object
		p, err := sonarr.New(pvr, libraryType)
		if err != nil {
			return nil, fmt.Errorf("failed initialising sonarr pvr %v: %w", pvr.Name, err)
		}

		return p, nil
	}

	return nil, errors.New("pvr not found")
}
