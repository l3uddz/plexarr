package main

import "strings"

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
