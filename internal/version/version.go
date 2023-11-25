// Package version is a singleton module which stores project build information.
package version

import (
	"fmt"
)

type buildInfo struct {
	buildNumber string
	buildDate   string
	buildCommit string
}

var bi buildInfo

func init() {
	valueIsNotAvailable := "N/A"

	bi = buildInfo{
		buildNumber: valueIsNotAvailable,
		buildDate:   valueIsNotAvailable,
		buildCommit: valueIsNotAvailable,
	}
}

// Set should be called from the main function to make application version details available for other app modules.
func Set(buildVersion, buildDate, buildCommit string) {
	if len(buildVersion) > 0 {
		bi.buildNumber = buildVersion
	}

	if len(buildDate) > 0 {
		bi.buildDate = buildDate
	}

	if len(buildCommit) > 0 {
		bi.buildCommit = buildCommit
	}
}

// BuildVersion sets version of the application.
func BuildVersion() string {
	return bi.buildNumber
}

// BuildDate sets date of the build.
func BuildDate() string {
	return bi.buildDate
}

// BuildCommit sets last commit id.
func BuildCommit() string {
	return bi.buildCommit
}

// Print - outputs build information right into terminal.
func Print() {
	fmt.Printf("Version:    %s\n", BuildVersion())
	fmt.Printf("Commit:     %s\n", BuildCommit())
	fmt.Printf("Build date: %s\n", BuildDate())
}
