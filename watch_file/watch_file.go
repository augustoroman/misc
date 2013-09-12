// watch_file is a tiny windows utility that forces NTFS to update the size of
// a file and then displays the file size.  It repeats this once a second until
// interrupted.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

var rate = flag.Duration("rate", time.Minute, "Rate at which the file is checked.  "+
	"Example rates: 1s (= once per sec), 1m (= once a minute), 2m45s, 500ms, 1h")

func doUpdate(target string) {
	if err := update(target); err != nil {
		fmt.Printf("Error trying to access [%s]: %v\n", target, err)
	} else if fi, err := os.Stat(target); err != nil {
		fmt.Printf("Error trying to access [%s]: %v\n", target, err)
	} else {
		fmt.Printf("%s: %14d\n", target, fi.Size())
	}
}

func usage() {
	fmt.Println("Usage is: watch_file [-rate=RATE] <filename>")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if len(flag.Args()) != 1 {
		usage()
		os.Exit(1)
	}
	target := flag.Args()[0]
	doUpdate(target)
	for _ = range time.Tick(*rate) {
		doUpdate(target)
	}
}
