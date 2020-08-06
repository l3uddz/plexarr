package main

import (
	"fmt"
	"github.com/l3uddz/plexarr"
	"github.com/l3uddz/plexarr/plex"
	"github.com/rs/zerolog/log"
	"time"
)

type plexLibraryItem struct {
	Name  string
	Type  plexarr.LibraryType
	Items []plex.MediaItem
}

func getPlexLibraryItems(p *plex.Client, libraries []string) ([]plexLibraryItem, error) {
	plexItems := make([]plexLibraryItem, 0)

	for _, library := range libraries {
		// get library items
		l := log.With().
			Str("library", library).
			Str("pvr", cli.PVR).
			Bool("dry_run", cli.DryRun).
			Logger()

		l.Debug().Msg("Retrieving plex library items...")

		items, libType, err := p.GetLibraryItems(library)
		if err != nil {
			return nil, fmt.Errorf("failed %q plex library items: %w", library, err)
		}

		if len(items) == 0 {
			return nil, fmt.Errorf("no plex library items found for: %v", library)
		}

		l.Info().
			Int("count", len(items)).
			Msg("Retrieved plex library items")

		plexItems = append(plexItems, plexLibraryItem{
			Name:  library,
			Type:  libType,
			Items: items,
		})
	}

	return plexItems, nil
}

func findDuplicateItems(items []plex.MediaItem) ([]plex.MediaItem, error) {
	// locate duplicate metadata_item ids
	duplicateMetadataItemIds := make([]plex.MediaItem, 0)
	uniqueMetadataItemIds := make(map[uint64]int)

	for _, item := range items {
		count, ok := uniqueMetadataItemIds[item.MetadataId]
		if !ok {
			// item does not exist
			uniqueMetadataItemIds[item.MetadataId] = 1
			continue
		}

		// the item was a duplicate
		if count == 1 {
			duplicateMetadataItemIds = append(duplicateMetadataItemIds, item)
		}

		uniqueMetadataItemIds[item.MetadataId]++
	}

	return duplicateMetadataItemIds, nil
}

func splitDuplicates(p *plex.Client, library plexLibraryItem) (int, error) {
	l := log.With().
		Str("library", library.Name).
		Str("pvr", cli.PVR).
		Bool("dry_run", cli.DryRun).
		Logger()

	duplicates, err := findDuplicateItems(library.Items)
	if err != nil {
		return 0, fmt.Errorf("failed finding duplicates items in plex library %q: %w", library.Name, err)
	}

	duplicatesSize := len(duplicates)

	// split duplicates
	splitSize := 0
	if len(duplicates) > 0 {
		l.Debug().
			Interface("duplicates", duplicates).
			Int("count", duplicatesSize).
			Msg("Duplicates found")
		l.Warn().
			Int("count", duplicatesSize).
			Msg("Duplicates found, splitting...")

		// iterate duplicates splitting
		for _, duplicate := range duplicates {
			if !cli.DryRun {
				err = p.Split(int(duplicate.MetadataId))
			} else {
				err = nil
			}

			if err != nil {
				return 0, fmt.Errorf("failed splitting duplicate item in plex library %q: %v: %w",
					library.Name, duplicate, err)
			}

			splitSize++
			l.Info().
				Str("path", duplicate.Path).
				Str("guid", duplicate.GUID).
				Uint64("metadata_item_id", duplicate.MetadataId).
				Msg("Split duplicate")

			if !cli.DryRun {
				time.Sleep(15 * time.Second)
			}
		}
	}

	return splitSize, nil
}
