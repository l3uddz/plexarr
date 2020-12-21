package main

import (
	"bufio"
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"os"
)

type updateFlag string

func (u updateFlag) Decode(ctx *kong.DecodeContext) error { return nil }
func (u updateFlag) IsBool() bool                         { return true }
func (u updateFlag) BeforeApply(app *kong.Kong, vars kong.Vars) error {
	// parse current version
	v, err := semver.Parse(Version)
	if err != nil {
		fmt.Printf("Failed parsing current build version: %v\n", err)
		app.Exit(1)
		return nil
	}

	// detect latest version
	fmt.Println("Checking for the latest version...")
	latest, found, err := selfupdate.DetectLatest("l3uddz/plexarr")
	if err != nil {
		fmt.Printf("Failed determining latest available version: %v\n", err)
		app.Exit(1)
		return nil
	}

	// check version
	if !found || latest.Version.LTE(v) {
		fmt.Printf("Already using the latest version: %v\n", Version)
		app.Exit(0)
		return nil
	}

	// ask update
	fmt.Printf("Do you want to update to the latest version: %v? (y/n): ", latest.Version)
	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil || (input != "y\n" && input != "n\n") {
		fmt.Println("Failed validating input...")
		app.Exit(1)
		return nil
	} else if input == "n\n" {
		app.Exit(0)
		return nil
	}

	// get existing executable path
	exe, err := os.Executable()
	if err != nil {
		fmt.Printf("Failed locating current executable path: %v\n", err)
		app.Exit(1)
		return nil
	}

	if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
		fmt.Printf("Failed updating existing binary to latest release: %v\n", err)
		app.Exit(1)
		return nil
	}

	fmt.Printf("Successfully updated to the latest version: %v\n", latest.Version)

	app.Exit(0)
	return nil
}
