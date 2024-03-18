package main;

import (
	"log"
	"os"
	"bit_tor/decode"
	"bit_tor/p2p"
	"fmt"
)

func main() {
	inPath := os.Args[1];
	// outPath := os.Args[2];
	tf, err := decode.Open(inPath);
	if err != nil {
		log.Fatal(err);
	}

	torrent, err := tf.GetTorrentFile();
	res, err := p2p.Download(torrent, 2);
	if err != nil {
		log.Fatal(err);
	}
	fmt.Println(res);
}
