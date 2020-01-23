package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/nakatamixi/go-ddls2m"
)

func main() {
	var (
		file      string
		withIndex bool
	)
	flags := flag.NewFlagSet("", flag.ContinueOnError)
	flags.StringVar(&file, "s", "", "spanner schema file")
	flags.BoolVar(&withIndex, "i", true, "with index")
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
	mbody, err := ddls2m.Convert(body, withIndex)
	if err != nil {
		panic(err)
	}
	fmt.Print(mbody)
}
func read(file string) (string, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	body := string(data)
	return body, nil

}
