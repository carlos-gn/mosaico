package main

import (
	"fmt"

	"github.com/progrium/darwinkit/macos/appkit"
)

func main() {
	workspace := appkit.Workspace_SharedWorkspace()
	apps := workspace.RunningApplications()

	for _, app := range apps {
		name := app.LocalizedName()
		bundleID := app.BundleIdentifier()
		fmt.Printf("%s (%s)'\n", name, bundleID)
	}
}
