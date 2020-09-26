package plex

import (
	"database/sql"
	"fmt"
	"github.com/l3uddz/plexarr"
	"net/url"
	"path/filepath"
	"strings"

	// database driver
	_ "github.com/mattn/go-sqlite3"
)

func newDatastore(path string) (*datastore, error) {
	q := url.Values{}
	q.Set("mode", "ro")

	db, err := sql.Open("sqlite3", plexarr.DSN(path, q))
	if err != nil {
		return nil, fmt.Errorf("could not open database: %v", err)
	}

	return &datastore{db: db}, nil
}

type datastore struct {
	db *sql.DB
}

type library struct {
	ID   int
	Name string
	Type plexarr.LibraryType
	Path string
}

func (d *datastore) Libraries() ([]library, error) {
	rows, err := d.db.Query(sqlSelectLibraries)
	if err != nil {
		return nil, fmt.Errorf("select libraries: %v", err)
	}

	defer rows.Close()

	libraries := make([]library, 0)
	for rows.Next() {
		l := library{}
		if err := rows.Scan(&l.ID, &l.Name, &l.Type, &l.Path); err != nil {
			return nil, fmt.Errorf("scan library row: %v", err)
		}

		libraries = append(libraries, l)
	}

	return libraries, nil
}

type MediaItem struct {
	LibraryId  uint64
	Path       string
	MetadataId uint64
	GUID       string
}

func (d *datastore) GetMediaItems(libraryId int) ([]MediaItem, error) {
	rows, err := d.db.Query(sqlSelectLibraryItemsMetadata, libraryId)
	if err != nil {
		return nil, fmt.Errorf("select media items: %v", err)
	}

	defer rows.Close()

	mediaItems := make([]MediaItem, 0)
	for rows.Next() {
		m := new(struct {
			LibraryId                                      *uint64
			LibraryName                                    *string
			SectionId                                      *uint64
			SectionPath                                    *string
			SectionDirectoryId                             *uint64
			SectionChildDirectoryId                        *uint64
			SectionChildDirectoryPath                      *string
			SectionChildDirectoryMetadataItemId            *uint64
			SectionChildDirectoryMetadataItemGuid          *string
			SectionChildDirectoryMetadataItemExternalGuids *string
		})
		if err := rows.Scan(&m.LibraryId, &m.LibraryName, &m.SectionId, &m.SectionPath, &m.SectionDirectoryId,
			&m.SectionChildDirectoryId, &m.SectionChildDirectoryPath, &m.SectionChildDirectoryMetadataItemId,
			&m.SectionChildDirectoryMetadataItemGuid, &m.SectionChildDirectoryMetadataItemExternalGuids); err != nil {
			return nil, fmt.Errorf("scan media item row: %v", err)
		}

		if m.LibraryId == nil || m.SectionPath == nil || m.SectionChildDirectoryPath == nil ||
			m.SectionChildDirectoryMetadataItemId == nil || m.SectionChildDirectoryMetadataItemGuid == nil {
			return nil, fmt.Errorf("invalid media item row: %v", m)
		}

		guid := ""
		if strings.HasPrefix(*m.SectionChildDirectoryMetadataItemGuid, "plex://") {
			// item has a plex guid - we are only able to handle this in specific scenarios
			if m.SectionChildDirectoryMetadataItemExternalGuids == nil {
				// no external guids were present ??
				return nil, fmt.Errorf("invalid media item row: %v", m)
			}

			guid = *m.SectionChildDirectoryMetadataItemExternalGuids
		} else {
			guid = *m.SectionChildDirectoryMetadataItemGuid
		}

		mediaItems = append(mediaItems, MediaItem{
			LibraryId:  *m.LibraryId,
			Path:       filepath.Join(*m.SectionPath, *m.SectionChildDirectoryPath),
			MetadataId: *m.SectionChildDirectoryMetadataItemId,
			GUID:       guid,
		})
	}

	return mediaItems, nil
}

//goland:noinspection ALL
const (
	sqlSelectLibraries = `
SELECT
    ls.id,
    ls.name,
    ls.section_type as type,
    sl.root_path
FROM
    library_sections ls
    JOIN section_locations sl ON sl.library_section_id = ls.id
`
	sqlSelectLibraryItemsMetadata = `
with ls as (
    SELECT
        ls.id AS library_id,
        ls.name AS library_name,
        sl.id AS section_id,
        sl.root_path AS section_root_path,
        d.id AS section_directory_id
    FROM
        library_sections ls
        JOIN section_locations sl ON sl.library_section_id = ls.id
        JOIN directories d ON d.library_section_id = ls.id
    WHERE
        d.parent_directory_id IS NULL
)
SELECT DISTINCT 
	ls.*,
    d.id AS child_directory_id,
    d.path AS child_directory_path,
    CASE
        WHEN mti3.guid IS NOT NULL THEN mti3.id
        WHEN mti2.guid IS NOT NULL THEN mti2.id
        WHEN mti.guid IS NOT NULL THEN mti.id
        ELSE NULL
    END AS child_directory_metadata_item_id,
    CASE
        WHEN mti3.guid IS NOT NULL THEN mti3.guid
        WHEN mti2.guid IS NOT NULL THEN mti2.guid
        WHEN mti.guid IS NOT NULL THEN mti.guid
        ELSE NULL
    END AS child_directory_metadata_item_guid
    , GROUP_CONCAT(t.tag) as child_directory_metadata_item_guids_external
FROM
    ls
    JOIN directories d ON d.parent_directory_id = ls.section_directory_id
    LEFT JOIN directories d2 ON d2.parent_directory_id = d.id
    JOIN media_parts mdp ON mdp.directory_id = d.id OR mdp.directory_id = d2.id
    JOIN media_items mdi ON mdi.id = mdp.media_item_id
    JOIN metadata_items mti ON mti.id = mdi.metadata_item_id
    LEFT JOIN metadata_items mti2 ON mti2.id = mti.parent_id
    LEFT JOIN metadata_items mti3 ON mti3.id = mti2.parent_id
	LEFT JOIN taggings tj ON tj.metadata_item_id = mti.id
    LEFT JOIN tags t ON t.id = tj.tag_id AND t.tag_type = 314
WHERE
   ls.library_id = $1
GROUP BY d.id
`
)
