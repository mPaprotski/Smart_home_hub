package main

import (
	"flag"
	"fmt"
	"github.com/mPaprotski/Smart_home_hub/internal/buildinfo"
	"os"
)

func main() {
	version := flag.Bool("version", false, "print scaffold version")
	flag.Parse()
	if *version {
		fmt.Println(buildinfo.Version, buildinfo.Commit)
		return
	}
	fmt.Fprintln(os.Stderr, "NOT_IMPLEMENTED: transport and device integration are future stages")
	os.Exit(2)
}
