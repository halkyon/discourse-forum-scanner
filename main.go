package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/halkyon/discourse-scanner/internal/postchecker"
)

const (
	displayVersionFlag = "version"
	discourseURLFlag   = "discourse-url"
	checkIntervalFlag  = "check-interval"
	filterKeywordsFlag = "filter-keywords"
)

var (
	version        = "dev"
	commit         = ""
	date           = ""
	displayVersion = flag.Bool(displayVersionFlag, false, "Display version information")
	discourseURL   = flag.String(discourseURLFlag, "", "URL to Discourse instance")
	checkInterval  = flag.Duration(checkIntervalFlag, 10*time.Minute, "Interval between getting posts")
	filterKeywords = flag.String(filterKeywordsFlag, "", "Comma separated list of keywords to filter posts by")
)

func validateFlags() error {
	if err := validateNotEmpty(discourseURLFlag, *discourseURL); err != nil {
		return err
	}
	if err := validateURL(discourseURLFlag, *discourseURL); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Printf("fatal: %+v\n", err)
		os.Exit(1)
	}
}

func run() error {
	flag.Parse()

	if *displayVersion {
		result := version
		if commit != "" {
			result = fmt.Sprintf("%s\ncommit: %s", result, commit)
		}
		if date != "" {
			result = fmt.Sprintf("%s\nbuilt at: %s", result, date)
		}
		result = fmt.Sprintf("%s\ngoos: %s\ngoarch: %s", result, runtime.GOOS, runtime.GOARCH)
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Sum != "" {
			result = fmt.Sprintf("%s\nmodule version: %s, checksum: %s", result, info.Main.Version, info.Main.Sum)
		}
		fmt.Println(result)
		return nil
	}

	if err := validateFlags(); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	done := make(chan error, 1)

	pc := postchecker.New(*discourseURL, *filterKeywords, *checkInterval)
	go pc.Run(ctx, done)

	if err := <-done; !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}
