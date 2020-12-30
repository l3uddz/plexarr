package plex

import (
	"fmt"
	"github.com/l3uddz/plexarr"
	"strings"
)

func (c *Client) GetLibraryItems(libraryName string) ([]MediaItem, plexarr.LibraryType, error) {
	// get library
	lib, err := c.getLibraryByName(libraryName)
	if err != nil {
		return nil, 0, err
	}

	// get library items
	items, err := c.store.GetMediaItems(lib.ID)
	if err != nil {
		return nil, 0, fmt.Errorf("retrieve library items: %v", err)
	}

	return items, lib.Type, nil
}

func (c *Client) getLibraryByName(name string) (*library, error) {
	for _, lib := range c.libraries {
		if strings.EqualFold(lib.Name, name) {
			return &lib, nil
		}
	}

	return nil, fmt.Errorf("no library found with name: %v", name)
}
