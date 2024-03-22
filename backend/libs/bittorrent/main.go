package main;

import (
	"os"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/libs/bittorrent/decode"
	"github.com/nikitaSstepanov/p2p-streaming-service/backend/libs/bittorrent/p2p"
	"fmt"
)

func main() {
	inPath := os.Args[1];
	tf, err := decode.Open(inPath);
	// fmt.Println(inPath);
	if err != nil {
		fmt.Println(err);
		return ;
	}
	torrent, err := tf.GetTorrentFile(); // -> Torrent
	// type Torrent struct {
		// PieceLength - вес одной части
		// Length - вес всего файла
		// Name - имя файла
		// PieceHashes - индекс части должен быть в диапазоне [0, len(PieceHashes)]
		// ... - остальное
	// }
	// по идее, больше тебе ничего от этой структуры не надо
	index := 2;
	res, err := p2p.Download(torrent, index); // ->Piece
	// type Piece struct {
		// index - номер части ( в 0-индексации)
		// buff - буфер (то, что скачал у других пиров)
		// ... - остальное
	// }
	if err != nil {
		fmt.Println(err);
		return ;
	}
	fmt.Println(res);
}
