package main

import (
	"log"
	"os"

	"github.com/Ehab-24/bittorrent/torrent"
)

func main() {
	log.SetFlags(0)

	helpString := "Usage:\tbit [inPath] [outPath]\n\t~ inPath:\tWhere to read the .torrent file from\n\t~ outPath:\tWhere to store the downloaded file"

	if len(os.Args) == 2 && (os.Args[1] == "--help" || os.Args[1] == "help") {
		log.Println(helpString)
		return
	}
	if len(os.Args) != 3 {
		log.Fatalf(helpString)
	}

	inPath := os.Args[1]
	outPath := os.Args[2]

	tf, err := torrent.Open(inPath)
	if err != nil {
		log.Fatal(err)
	}

	if err = tf.DownloadToFile(outPath); err != nil {
		log.Fatal(err)
	}
	log.Println("Download complete")
}
