package decode;

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"os"
	"net"
	"net/url"
	"net/http"
	"strconv"
	"encoding/binary"
	// "io/ioutil"
	// "bit_tor/peers"
	"github.com/jackpal/bencode-go"
)

type BencodeTrackerResponse struct {
	Peers 		string 	`bencode: "peers"`
	Interval 	int 	`bencode: "interval"`
}

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type bencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

type Peer struct {
	Ip 		net.IP
	Port 	uint16
}

type Torrent struct {
	Peers 		[]Peer
	PeerID 		[20]byte
	InfoHash 	[20]byte
	PieceHashes [][20]byte
	PieceLength int 
	Length 		int
	Name 		string
}

func (p *Peer) String() string {
	var res string;
	res += p.Ip.String();
	res += ":"
	res += strconv.Itoa(int(p.Port));
	return res;
}

func Reverse(s string) string {
    runes := []rune(s)
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes)
}

func (t * TorrentFile) buildTrackerUrl(peerID [20]byte) (string, error) {
	// fmt.Printf("%s\n", t.Announce);
	// request := fmt.Sprintf("%s?info_hash=%s&peer_id=%s&port=%d&compact=1&uploaded=0&downloaded=0&left=%d", 
	// 	t.Announce, string(t.InfoHash[:]), string(peerID[:]), Port, t.Length);
	base, err := url.Parse(t.Announce);
	// p, _ := strconv.ParseInt(t.Announce, 64, 10);
	// Port := int(p);
	p := ""
	flag := false;
	for i := len(t.Announce)-1; i >= 0; i-- {
		if flag && t.Announce[i] == ':' {
			break;
		}
		if flag {
			p += string(t.Announce[i]);	
		}
		if !flag && t.Announce[i] == '/' {
			flag = true
		}
	}
	Port := Reverse(p);
	// fmt.Println(Port);
	if err != nil {
		return "", err;
	}
	params := url.Values {
		"info_hash" : []string{string(t.InfoHash[:])},
		"peer_id": []string{string(peerID[:])},
		"port": []string{Port},
		"compact": []string{"1"},
		"uploaded": []string{"0"},
		"downloaded":[]string{"0"},
		"left": []string{strconv.Itoa(t.Length)},
	};
	base.RawQuery = params.Encode();
	return base.String(), nil;
}

func (t *TorrentFile) UnmarshalPeers(Peers []byte) ([]Peer, error) {
	var (
		size = 6
		countPeers = len(Peers) / 6
	)
	if len(Peers) % countPeers != 0 {
		return []Peer{}, fmt.Errorf("received malformed peers");
	}
	peers := make([]Peer, countPeers);
	for i := 0; i < countPeers; i++ {
		offset := i * size;
		peers[i].Ip = net.IP(Peers[offset : offset + 4]);
		peers[i].Port = binary.BigEndian.Uint16([]byte(Peers[offset +4 : offset + 6]));
	}
	return peers, nil;
}

func (t *TorrentFile) requestPeers(peerID [20]byte) ([]Peer, error) {
	request, err := t.buildTrackerUrl(peerID);
	if err != 	nil {
		return []Peer{}, err;
	}

	req, err := http.NewRequest(http.MethodGet, request, nil);

	if err != nil {
		return []Peer{}, err;
	}
	res, err := http.DefaultClient.Do(req);
	
	if err != nil {
		return []Peer{}, err;
	}
	// body, err := ioutil.ReadAll(res.Body);
	// if err != nil {					
	// 	return []string{" "}, err;
	// }
	// fmt.Println(body);
	trackerResp := BencodeTrackerResponse{};
	err = bencode.Unmarshal(res.Body, &trackerResp);
	if err != nil {
		return []Peer{}, err;
	}
	peers, err := t.UnmarshalPeers([]byte(trackerResp.Peers));
	if err != nil {
		return []Peer{}, err;
	}
	// fmt.Println(peers);
	return peers, nil;

}

func (t *TorrentFile) GetTorrentFile() (Torrent, error) {
	var peerID [20]byte;
	_, err := rand.Read(peerID[:]);
	if err != nil {
		return Torrent{}, err;
	}
	peers, err := t.requestPeers(peerID);
	if err != nil {
		fmt.Println(err);
		return Torrent{}, err
	}
	return Torrent {
		Peers: peers,
		PeerID: peerID,
		InfoHash: t.InfoHash,
		PieceHashes: t.PieceHashes,
		PieceLength: t.PieceLength,
		Length: t.Length,
		Name: t.Name,
	}, nil;
}

func Open(path string) (TorrentFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return TorrentFile{}, err
	}
	defer file.Close()

	bto := bencodeTorrent{}
	err = bencode.Unmarshal(file, &bto)
	if err != nil {
		return TorrentFile{}, err
	}
	return bto.toTorrentFile()
}

func (i *bencodeInfo) hash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, *i)
	if err != nil {
		return [20]byte{}, err
	}
	h := sha1.Sum(buf.Bytes())
	return h, nil
}

func (i *bencodeInfo) splitPieceHashes() ([][20]byte, error) {
	hashLen := 20
	buf := []byte(i.Pieces)
	if len(buf) % hashLen != 0 {
		return nil, fmt.Errorf("Received malformed pieces of length %d", len(buf))
	}
	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}
	return hashes, nil
}

func (bto *bencodeTorrent) toTorrentFile() (TorrentFile, error) {
	infoHash, err := bto.Info.hash()
	if err != nil {
		return TorrentFile{}, err
	}
	pieceHashes, err := bto.Info.splitPieceHashes()
	if err != nil {
		return TorrentFile{}, err
	}
	t := TorrentFile{
		Announce:    bto.Announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: bto.Info.PieceLength,
		Length:      bto.Info.Length,
		Name:        bto.Info.Name,
	}
	return t, nil
}


