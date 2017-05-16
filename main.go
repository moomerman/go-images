package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"mime"
	"os"

	"github.com/moomerman/go-images/application"
)

// Version is set automatically at build time
var Version = "Development"

// Options holds the configuration for the Application
type Options struct {
	bindAddress string
	configPath  string
}

func main() {
	opts, err := parseArgs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// read configuration file
	data, err := ioutil.ReadFile(opts.configPath)
	if err != nil {
		fmt.Printf("Failed to read configuration file %s: %v\n", opts.configPath, err)
		os.Exit(1)
	}

	// parse config
	config, err := application.NewConfig(data)
	if err != nil {
		fmt.Printf("Failed to read config file %s: %v\n", opts.configPath, err)
		os.Exit(1)
	}

	mime.AddExtensionType(".gdoc", "application/vnd.google-apps.document")
	mime.AddExtensionType(".gslide", "application/vnd.google-apps.presentation")
	mime.AddExtensionType(".gsheet", "application/vnd.google-apps.spreadsheet")
	mime.AddExtensionType(".gdraw", "application/vnd.google-apps.drawing")
	mime.AddExtensionType(".gform", "application/vnd.google-apps.form")
	mime.AddExtensionType(".odt", "application/vnd.oasis.opendocument.text")
	mime.AddExtensionType(".oxps", "application/oxps")
	mime.AddExtensionType(".odp", "application/vnd.oasis.opendocument.presentation")
	mime.AddExtensionType(".oxt", "application/vnd.openofficeorg.extension")
	mime.AddExtensionType(".ods", "application/vnd.oasis.opendocument.spreadsheet")
	mime.AddExtensionType(".pub", "application/x-mspublisher")
	mime.AddExtensionType(".wps", "application/vnd.ms-works")
	mime.AddExtensionType(".key", "application/x-iwork-keynote-sffkey")
	mime.AddExtensionType(".pages", "application/x-iwork-pages-sffpages")
	mime.AddExtensionType(".numbers", "application/x-iwork-numbers-sffnumbers")

	// run application
	config.Bind = opts.bindAddress
	config.Version = Version
	application.Run(config)
}

func parseArgs() (*Options, error) {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-bind <addr>] <config_file>\n\n", os.Args[0])
	}
	bindAddress := flag.String("bind", "localhost:3000", "bind address")
	flag.Parse()

	if len(flag.Args()) != 1 {
		flag.Usage()
		return nil, fmt.Errorf("No config file path specified.")
	}

	return &Options{
		bindAddress: *bindAddress,
		configPath:  flag.Arg(0),
	}, nil

}
