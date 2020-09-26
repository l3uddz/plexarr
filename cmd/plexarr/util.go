package main

import (
	"fmt"
	"strings"
)

func getPreferredGuid(guids []string) string {
	for _, guid := range guids {
		if strings.Contains(guid, "tvdb") {
			return guid
		} else if strings.Contains(guid, "imdb") {
			return guid
		}
	}

	return guids[0]
}

func guidsMatched(plexGuids []string, pvrGuids []string) bool {
	for _, pvrGuid := range pvrGuids {
		for _, plexGuid := range plexGuids {
			if strings.HasPrefix(plexGuid, pvrGuid) {
				return true
			}
		}
	}
	return false
}

func getPlexGuids(externalGuids string) ([]string, error) {
	guids := strings.Split(externalGuids, ",")
	formattedGuids := make([]string, 0)

	for _, guid := range guids {
		if strings.HasPrefix(guid, "tvdb://") {
			formattedGuids = append(formattedGuids, fmt.Sprintf("com.plexapp.agents.the%s", guid))
			continue
		} else if strings.HasPrefix(guid, "com.plexapp.agents") {
			formattedGuids = append(formattedGuids, guid)
			continue
		}

		formattedGuids = append(formattedGuids, fmt.Sprintf("com.plexapp.agents.%s", guid))
	}

	if len(formattedGuids) == 0 {
		return nil, fmt.Errorf("unable to format guids: %v", externalGuids)
	}

	return formattedGuids, nil
}
