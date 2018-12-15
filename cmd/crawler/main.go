package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/kouheiszk/ig-crawler"
	"log"
	"os"
	"os/signal"
)

const (
	Name    = "crawler"
	Version = "0.0.1"
)

type CommandLineOptions struct {
	Type        string `short:"t" long:"type" description:"profile | posts" default:"profile"`
	Username    string `short:"u" long:"username" description:"Target username." required:"true"`
	Concurrency int    `short:"c" long:"concurrency" description:"Concurrency number of converting images to pdf." default:"2"`
	Version     bool   `short:"V" long:"version" description:"Displays version information."`
}

func main() {
	// -----------------------------------------------------------------------------------
	// Handle SIGINT (Ctrl + C)
	// -----------------------------------------------------------------------------------

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	go func() {
		<-signalChan
		fmt.Println("Operation has been aborted.")
		os.Exit(2)
	}()

	// -----------------------------------------------------------------------------------
	// Parse arguments
	// -----------------------------------------------------------------------------------

	var opts CommandLineOptions
	_, err := flags.NewParser(&opts, flags.HelpFlag|flags.PassDoubleDash).Parse()
	if err != nil {
		log.Fatalln(err)
	}

	// -----------------------------------------------------------------------------------
	// Handle version command
	// -----------------------------------------------------------------------------------

	if opts.Version {
		fmt.Println(Version)
		return
	}

	switch opts.Type {
	case "profile":
		url, err := crawler.FetchProfileImage(&crawler.Config{
			Username: opts.Username,
		})
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(url)
	case "posts":
		fmt.Println("posts")
	default:
		log.Fatalln(fmt.Errorf("invalid type: %s", opts.Type))
	}
}
