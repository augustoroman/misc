// watch_file is a tiny windows utility that forces NTFS to update the size of
// a file and then displays the file size.  It repeats this once a second until
// interrupted.
package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage is: watch_file [filename]")
		os.Exit(0)
	}
	target := os.Args[1]
	for _ = range time.Tick(1 * time.Second) {
		if err := update(target); err != nil {
			fmt.Printf("Error trying to access [%s]: %v\n", target, err)
		} else if fi, err := os.Stat(target); err != nil {
			fmt.Printf("Error trying to access [%s]: %v\n", target, err)
		} else {
			fmt.Printf("%s: %14d\n", target, fi.Size())
		}
	}
}
