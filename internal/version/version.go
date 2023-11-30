// Package version is a singleton module which stores project build information.
package version

import (
	"fmt"
)

type buildInfo struct {
	number     string
	date       string
	commitHash string
}

var bi buildInfo

func init() {
	valueIsNotAvailable := "N/A"

	bi = buildInfo{
		number:     valueIsNotAvailable,
		date:       valueIsNotAvailable,
		commitHash: valueIsNotAvailable,
	}
}

// Set should be called from the main function to make application version details available for other app modules.
func Set(buildVersion, buildDate, buildCommit string) {
	if len(buildVersion) > 0 {
		bi.number = buildVersion
	}

	if len(buildDate) > 0 {
		bi.date = buildDate
	}

	if len(buildCommit) > 0 {
		bi.commitHash = buildCommit
	}
}

// Number sets version of the application.
func Number() string {
	return bi.number
}

// BuildDate sets date of the build.
func BuildDate() string {
	return bi.date
}

// CommitHash sets last commit id.
func CommitHash() string {
	return bi.commitHash
}

// Print - outputs build information right into terminal.
func Print() {
	fmt.Printf("Version:    %s\n", Number())
	fmt.Printf("Commit:     %s\n", CommitHash())
	fmt.Printf("Build date: %s\n", BuildDate())
}
