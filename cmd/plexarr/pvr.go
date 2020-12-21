package main

import (
	"errors"
	"fmt"
	"github.com/l3uddz/plexarr"
	"github.com/l3uddz/plexarr/pvrs/radarr"
	"github.com/l3uddz/plexarr/pvrs/sonarr"
	"github.com/rs/zerolog/log"
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
			if lib.Type != plexarr.TvLibrary {
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

func getPvrItems(names []string, cfg config, plexItems []plexLibraryItem) (map[string]plexarr.PvrItem, error) {
	pvrItems := make(map[string]plexarr.PvrItem)

	// iterate pvr names
	for _, pvrName := range names {
		// get pvr object
		pvr, err := getPvr(pvrName, cfg, plexItems)
		if err != nil {
			return nil, fmt.Errorf("initialise pvr: %v: %w", pvrName, err)
		}

		// retrieve pvr items
		items, err := pvr.GetLibraryItems()
		if err != nil {
			return nil, fmt.Errorf("retrieve pvr library items: %v: %w", pvrName, err)
		} else if len(items) == 0 {
			return nil, fmt.Errorf("retrieve pvr library items: %v: no items found", pvrName)
		}

		itemsSkipped := 0
		itemsAdded := 0

		pl := log.With().
			Str("pvr", pvrName).
			Logger()

		// process pvr items
		for key, item := range items {
			// does key already exist in pvrItems (have we seen this path before?)
			if _, exists := pvrItems[key]; exists {
				// this key (path) already exists, ignore it
				pl.Warn().
					Interface("item", item).
					Msg("Path is not unique to this pvr, skipping item(s)")

				itemsSkipped++
				delete(pvrItems, key)
				continue
			}

			itemsAdded++
			pvrItems[key] = item
		}

		pl.Info().
			Int("skipped", itemsSkipped).
			Int("added", itemsAdded).
			Msg("Retrieved pvr library items")
	}

	return pvrItems, nil
}
