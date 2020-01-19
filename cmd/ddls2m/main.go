package main

import (
	"flag"
	"io/ioutil"
	"os"

	"github.com/nakatamixi/go-ddls2m"
)

func main() {
	var (
		file  string
		debug bool
	)
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	flags.StringVar(&file, "s", "", "spanner schema file")
	flags.BoolVar(&debug, "d", false, "debug print")
	if err := flags.Parse(os.Args[1:]); err != nil {
		flags.Usage()
		return
	}
	if file == "" {
		flags.Usage()
		return
	}
	body, err := read(file)
	if err != nil {
		panic(err)
	}
	ddls2m.Convert(body, debug)
}
func read(file string) (string, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	body := string(data)
	return body, nil

}
