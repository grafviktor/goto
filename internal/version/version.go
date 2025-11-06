// Package version is a singleton module which stores project build information.
package version

import (
	"fmt"
)

type buildInfo struct {
	number      string
	date        string
	commitHash  string
	buildBranch string
}

var bi buildInfo

func init() { //nolint:gochecknoinits // To set default values.
	valueIsNotAvailable := "N/A"

	bi = buildInfo{
		number:      valueIsNotAvailable,
		date:        valueIsNotAvailable,
		commitHash:  valueIsNotAvailable,
		buildBranch: valueIsNotAvailable,
	}
}

// Set should be called from the main function to make application version details available for other app modules.
func Set(buildVersion, buildCommit, buildBranch, buildDate string) {
	if len(buildVersion) > 0 {
		bi.number = buildVersion
	}

	if len(buildDate) > 0 {
		bi.date = buildDate
	}

	if len(buildCommit) > 0 {
		bi.commitHash = buildCommit
	}

	if len(buildBranch) > 0 {
		bi.buildBranch = buildBranch
	}
}

// Number returns version of the application.
func Number() string {
	return bi.number
}

// BuildDate returns date of the build.
func BuildDate() string {
	return bi.date
}

// BuildBranch returns branch which was used to build the app.
func BuildBranch() string {
	return bi.buildBranch
}

// CommitHash returns last commit id which was used to build the app.
func CommitHash() string {
	return bi.commitHash
}

// Print - outputs build information right into the terminal.
func Print() {
	fmt.Printf("Version:    %s\n", Number())
	fmt.Printf("Commit:     %s\n", CommitHash())
	fmt.Printf("Branch:     %s\n", BuildBranch())
	fmt.Printf("Build date: %s\n", BuildDate())
	fmt.Println()
}

type loggerInterface interface {
	Info(format string, args ...any)
}

// Print - outputs build information right into the terminal.
func LogDetails(logger loggerInterface) {
	logger.Info("[MAIN] Version:    %s", Number())
	logger.Info("[MAIN] Commit:     %s", CommitHash())
	logger.Info("[MAIN] Branch:     %s", BuildBranch())
	logger.Info("[MAIN] Build date: %s", BuildDate())
}
