/*
This program provides a formatter for ini files.
*/
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"unicode/utf8"

	ini "github.com/pierrec/go-ini"
)

func main() {
	var outName, comment, sliceSep string
	flag.StringVar(&outName, "w", "", "write result to file instead of stdout")
	flag.StringVar(&comment, "c", string(ini.DefaultComment), "comment character")
	flag.StringVar(&sliceSep, "sep", string(ini.DefaultSliceSeparator), "comment character")
	var sensitive, merge bool
	flag.BoolVar(&sensitive, "s", false, "make section and key names case sensitive")
	flag.BoolVar(&merge, "m", false, "merge sections with same name")
	flag.Usage = usage
	flag.Parse()

	var output io.Writer
	if outName == "" {
		output = os.Stdout
	} else {
		f, err := os.Create(outName)
		if err != nil {
			log.Println(err)
			return
		}
		defer f.Close()
		output = f
	}

	var input io.Reader
	if args := flag.Args(); len(args) == 0 {
		input = os.Stdin
	} else {
		fname := args[0]
		f, err := os.Open(fname)
		if err != nil {
			log.Println(err)
			return
		}
		defer f.Close()
		input = f
	}

	var options []ini.Option
	sep, _ := utf8.DecodeRuneInString(sliceSep)
	options = append(options, ini.SliceSeparator(sep))
	if comment != "" {
		options = append(options, ini.Comment(comment))
	}
	if sensitive {
		options = append(options, ini.CaseSensitive())
	}
	if merge {
		options = append(options, ini.MergeSections())
	}

	conf, _ := ini.New(options...)
	_, err := conf.ReadFrom(input)
	if err != nil {
		log.Println(err)
		return
	}

	_, err = conf.WriteTo(output)
	if err != nil {
		log.Println(err)
		return
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s [flags] [path ...]\n", os.Args[0])
	flag.PrintDefaults()
}
