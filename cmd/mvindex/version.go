// Copyright (c) 2020-2024 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package main

import (
	"fmt"
	"runtime"
)

var (
	company           = "Mavryk Dynamics LTD."
	envPrefix         = "MV"
	appName           = "mvindex"
	apiVersion        = "v018-2024-03-26"
	version    string = "v18.0"
	commit     string = "dev"
)

func UserAgent() string {
	return fmt.Sprintf("%s/%s.%s",
		appName,
		version,
		commit,
	)
}

func printVersion() {
	fmt.Printf("Mavryk L1 Indexer by %s\n", company)
	fmt.Printf("Version: %s (%s)\n", version, commit)
	fmt.Printf("API version: %s\n", apiVersion)
	fmt.Printf("Go version: %s\n", runtime.Version())
}
