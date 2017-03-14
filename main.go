package main

import (
	"flag"

	"github.com/slavikmanukyan/itm/fs"
)

func logError(err string) {

}

func main() {
	source := flag.String("source", "", "source directory")
	destination := flag.String("destination", "", "backup destination")
	// recover = := flag.Bool("recover", false, "recover files")

	flag.Parse()

	if (*source != "" && *destination == "") || (*source == "" && *destination != "") {
		logError("Wrong args")
		return
	}

	if *destination != "" {
		fs.CopyDir(*source, *destination)
	}

}
