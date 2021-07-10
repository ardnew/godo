package main

import (
	"flag"
	"fmt"
)

var (
	PROJECT   string
	IMPORT    string
	VERSION   string
	BUILDTIME string
	PLATFORM  string
	BRANCH    string
	REVISION  string
)

func shortVersion() string {
	return fmt.Sprintf("%s %s", PROJECT, VERSION)
}

func longVersion() string {
	build := BUILDTIME
	if "" != BRANCH && "" != REVISION {
		build = fmt.Sprintf("(%s@%s) %s", BRANCH, REVISION, build)
	}
	return fmt.Sprintf("%s %s %s", shortVersion(), PLATFORM, build)
}

func main() {

	var (
		argVersion bool
	)

	flag.BoolVar(&argVersion, "v", false, "Display version information")
	flag.Parse()

	if argVersion {
		fmt.Println(longVersion())
	} else {
		// main
	}
}
