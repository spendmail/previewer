package main

import (
	"flag"
	"log"

	internalconfig "github.com/spendmail/previewer/internal/config"
	internalresizer "github.com/spendmail/previewer/internal/resizer"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "/etc/previewer/previewer.toml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	// Config initialization.
	_, err := internalconfig.NewConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	width := uint(480)
	height := uint(320)
	inputFilename := "/home/spendlively/vhosts/previewer/help/images/me.jpg"
	outputFilename := "/home/spendlively/vhosts/previewer/help/images/me_result.jpg"

	resizer := internalresizer.New()
	err = resizer.Resize(width, height, inputFilename, outputFilename)
	if err != nil {
		panic(err)
	}
}