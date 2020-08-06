package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/l3uddz/plexarr"
	"github.com/l3uddz/plexarr/plex"
	"github.com/l3uddz/plexarr/pvrs/radarr"
	"github.com/l3uddz/plexarr/pvrs/sonarr"
	"github.com/natefinch/lumberjack"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type config struct {
	Plex plex.Config `yaml:"plex"`

	// PVRs
	Pvr struct {
		Radarr []radarr.Config `yaml:"radarr"`
		Sonarr []sonarr.Config `yaml:"sonarr"`
	} `yaml:"pvr"`
}

var (
	// Release variables
	Version   string
	Timestamp string
	GitCommit string

	// CLI
	cli struct {
		globals

		// flags
		PVR     string   `required:"1" type:"string" env:"PLEXARR_PVR" help:"PVR to match from"`
		Library []string `required:"1" type:"string" env:"PLEXARR_LIBRARY" help:"Plex Library to match against"`

		Config    string `type:"path" default:"${config_file}" env:"PLEXARR_CONFIG" help:"Config file path"`
		Log       string `type:"path" default:"${log_file}" env:"PLEXARR_LOG" help:"Log file path"`
		Verbosity int    `type:"counter" default:"0" short:"v" env:"PLEXARR_VERBOSITY" help:"Log level verbosity"`

		DryRun bool `type:"bool" default:"0" env:"PLEXARR_DRY_RUN" help:"Dry run mode"`
	}
)

type globals struct {
	Version versionFlag `name:"version" help:"Print version information and quit"`
}

type versionFlag string

func (v versionFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (v versionFlag) IsBool() bool                         { return true }
func (v versionFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	fmt.Println(vars["version"])
	app.Exit(0)
	return nil
}

func main() {
	// parse cli
	ctx := kong.Parse(&cli,
		kong.Name("plexarr"),
		kong.Description("Fix mismatched media in Plex mastered by Sonarr/Radarr"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Summary: true,
			Compact: true,
		}),
		kong.Vars{
			"version":     fmt.Sprintf("%s (%s@%s)", Version, GitCommit, Timestamp),
			"config_file": filepath.Join(defaultConfigPath(), "config.yml"),
			"log_file":    filepath.Join(defaultConfigPath(), "activity.log"),
		},
	)

	if err := ctx.Validate(); err != nil {
		fmt.Println("Failed parsing cli:", err)
		return
	}

	// logger
	logger := log.Output(io.MultiWriter(zerolog.ConsoleWriter{
		Out: os.Stderr,
	}, zerolog.ConsoleWriter{
		Out: &lumberjack.Logger{
			Filename:   cli.Log,
			MaxSize:    5,
			MaxAge:     14,
			MaxBackups: 5,
		},
		NoColor: true,
	}))

	switch {
	case cli.Verbosity == 1:
		log.Logger = logger.Level(zerolog.DebugLevel)
	case cli.Verbosity > 1:
		log.Logger = logger.Level(zerolog.TraceLevel)
	default:
		log.Logger = logger.Level(zerolog.InfoLevel)
	}

	// config
	file, err := os.Open(cli.Config)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed opening config")
	}
	defer file.Close()

	cfg := config{}
	decoder := yaml.NewDecoder(file)
	decoder.SetStrict(true)
	err = decoder.Decode(&cfg)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed decoding config")
	}

	switch {
	case cfg.Plex.URL == "":
		log.Fatal().Msg("You must set a plex url in your configuration")
	case cfg.Plex.Token == "":
		log.Fatal().Msg("You must set a plex token in your configuration")
	case cfg.Plex.Database == "":
		log.Fatal().Msg("You must set a plex database in your configuration")
	}

	// plex
	p, err := plex.New(cfg.Plex)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed initialising plex")
	}

	if err := p.Available(); err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed validating plex availability")
	}

	// get library items
	l := log.With().
		Str("pvr", cli.PVR).
		Bool("dry_run", cli.DryRun).
		Logger()

	plexItems, err := getPlexLibraryItems(p, cli.Library)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed retrieving items from plex libraries")
	}

	// find and split duplicate items
	l.Debug().Msg("Checking for duplicates...")

	splitSize := 0
	for _, lib := range plexItems {
		split, err := splitDuplicates(p, lib)
		if err != nil {
			log.Fatal().
				Err(err).
				Msg("Failed finding and splitting duplicate plex library items")
		}
		splitSize += split
	}

	// refresh items (post-split)
	if splitSize > 0 {
		l.Info().
			Int("count", splitSize).
			Msg("Finished splitting all duplicate items")

		time.Sleep(10 * time.Second)
		l.Info().Msg("Refreshing plex library items...")

		plexItems, err = getPlexLibraryItems(p, cli.Library)
		if err != nil {
			log.Fatal().
				Err(err).
				Msg("Failed retrieving items from plex libraries post-split")
		}
	} else {
		l.Info().Msg("No duplicates found!")
	}

	// retrieve items from pvr
	pvr, err := getPvr(cli.PVR, cfg, plexItems[0].Type)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed initialising pvr")
	}

	pvrItems, err := pvr.GetLibraryItems()
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Failed retrieving pvr library items")
	}

	if len(pvrItems) == 0 {
		l.Fatal().Msg("No pvr library items retrieved?")
	}

	l.Info().
		Int("count", len(pvrItems)).
		Msg("Retrieved pvr library items")

	// track items not matched (display debug log)
	plexItemsNotFound := make([]plex.MediaItem, 0)
	pvrItemsNotFound := make(map[string]plexarr.PvrItem)
	for k, v := range pvrItems {
		pvrItemsNotFound[k] = v
	}

	// iterate plex items matching to pvr items
	itemsToFix := make(map[plex.MediaItem]plexarr.PvrItem)

	for _, plexLibrary := range plexItems {
		for _, plexItem := range plexLibrary.Items {
			// plex item found in pvr items?
			pvrItem, ok := pvrItems[plexItem.Path]
			if !ok {
				// this plex item not found in pvr
				plexItemsNotFound = append(plexItemsNotFound, plexItem)
				continue
			} else {
				// plex item found in pvr
				delete(pvrItemsNotFound, plexItem.Path)
			}

			// validate match against pvr item
			validMatch := false
			for _, guid := range pvrItem.GUID {
				if strings.HasPrefix(plexItem.GUID, guid) {
					validMatch = true
					break
				}
			}

			if validMatch {
				l.Trace().
					Interface("plex_item", plexItem).
					Interface("pvr_item", pvrItem).
					Msg("Match validated")
				continue
			}

			// store item to fix
			itemsToFix[plexItem] = pvrItem
		}
	}

	// display files in pvr / plex that cannot be matched
	defer func(missingPlexItems []plex.MediaItem, missingPvrItems map[string]plexarr.PvrItem) {
		// show missing plex items
		if len(missingPlexItems) > 0 {
			for _, plexItem := range missingPlexItems {
				log.Debug().
					Interface("plex_item", plexItem).
					Msg("Cannot match plex library item to pvr item...")
			}
		}

		// show missing pvr items
		if len(missingPvrItems) > 0 {
			for _, pvrItem := range missingPvrItems {
				log.Debug().
					Interface("pvr_item", pvrItem).
					Msg("Cannot match pvr item to plex library item...")
			}
		}

	}(plexItemsNotFound, pvrItemsNotFound)

	// proceed no further if no mismatches
	if len(itemsToFix) == 0 {
		log.Info().Msg("No mismatched items found!")
		return
	}

	l.Info().
		Int("count", len(itemsToFix)).
		Msg("Mismatched found, fixing...")

	// fix matches
	fixedSize := 0
	for plexItem, pvrItem := range itemsToFix {
		newGuid := fmt.Sprintf("%s?lang=en", getPreferredGuid(pvrItem.GUID))

		l.Debug().
			Str("plex_path", plexItem.Path).
			Str("plex_guid", plexItem.GUID).
			Str("pvr_path", pvrItem.PvrPath).
			Str("pvr_guid", newGuid).
			Msgf("Fixing match to %v", newGuid)

		if !cli.DryRun {
			err = p.Match(int(plexItem.MetadataId), pvrItem.Title, newGuid)
		} else {
			err = nil
		}

		if err != nil {
			log.Fatal().
				Err(err).
				Interface("plex_item", plexItem).
				Interface("pvr_item", pvrItem).
				Msg("Failed fixing match")
		}

		fixedSize++
		log.Info().
			Str("plex_path", plexItem.Path).
			Str("plex_guid", plexItem.GUID).
			Str("pvr_path", pvrItem.PvrPath).
			Str("pvr_guid", newGuid).
			Msg("Fixed match")

		if !cli.DryRun {
			time.Sleep(15 * time.Second)
		}
	}

	if fixedSize > 0 {
		l.Info().
			Int("count", fixedSize).
			Msg("Finished fixing matches")
	}

	log.Info().Msg("Finished!")
}
