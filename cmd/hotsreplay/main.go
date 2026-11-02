package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/hotsbuff/stormparser-go/replay"
)

func main() {
	showSummary := flag.Bool("summary", false, "print typed match summary")
	showHeader := flag.Bool("header", false, "print the complete decoded header")
	showDetails := flag.Bool("details", false, "print replay.details")
	showInitData := flag.Bool("initdata", false, "print replay.initData")
	showAttributes := flag.Bool("attributes", false, "print replay.attributes.events")
	showGameEvents := flag.Bool("gameevents", false, "print replay.game.events")
	showMessageEvents := flag.Bool("messageevents", false, "print replay.message.events")
	showTrackerEvents := flag.Bool("trackerevents", false, "print replay.tracker.events")
	extract := flag.String("extract", "", "write an archive file to standard output")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: hotsreplay [options] replay.StormReplay")
		os.Exit(2)
	}

	r, err := replay.Open(flag.Arg(0))
	if err != nil {
		fatal(err)
	}
	defer r.Close()

	if *extract != "" {
		data, err := r.ReadFile(*extract)
		if err != nil {
			fatal(err)
		}
		if _, err := os.Stdout.Write(data); err != nil {
			fatal(err)
		}
		return
	}
	if *showSummary {
		summary, err := r.Summary()
		if err != nil {
			fatal(err)
		}
		printJSON(summary)
		return
	}

	anyOutput := *showHeader || *showDetails || *showInitData || *showAttributes || *showGameEvents || *showMessageEvents || *showTrackerEvents
	if !anyOutput {
		printJSON(map[string]any{
			"version":        r.Version(),
			"protocol_build": r.Schema.Build,
			"exact_protocol": r.Schema.Exact,
		})
		return
	}
	output := make(map[string]any)
	if *showHeader {
		output["header"] = r.Header
	}
	if *showDetails {
		output["details"], err = r.Details()
		if err != nil {
			fatal(err)
		}
	}
	if *showInitData {
		output["init_data"], err = r.InitData()
		if err != nil {
			fatal(err)
		}
	}
	if *showAttributes {
		output["attributes"], err = r.Attributes()
		if err != nil {
			fatal(err)
		}
	}
	if *showGameEvents {
		output["game_events"], err = r.GameEvents()
		if err != nil {
			fatal(err)
		}
	}
	if *showMessageEvents {
		output["message_events"], err = r.MessageEvents()
		if err != nil {
			fatal(err)
		}
	}
	if *showTrackerEvents {
		output["tracker_events"], err = r.TrackerEvents()
		if err != nil {
			fatal(err)
		}
	}
	printJSON(output)
}

func printJSON(value any) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
