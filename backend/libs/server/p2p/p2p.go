package p2p;

import (
	"net"
	"bit_tor/decode"
	"time"
	"encoding/binary"
	bt "bit_tor/BitField"
	"io"
	"fmt"
	"bytes"
)

const (
	Choke = iota
	Unchoke
	Interested
	NotInterested
	Have
	bitF
	Req
	Pic
	Cancel
)

type MSG struct {
	ID 		int
	Payload []byte
}

type Client struct {
	conn 		net.Conn
	InfoHash 	[20]byte
	PeerId 		[20]byte
	Peer 		decode.Peer
	bt_field	bt.BT
	Choked		bool
}

type Piece struct {
	begin 	int
	end 	int
	index	int
	buff 	[]byte
	hash	[20]byte
}

type ProgressInfo struct {
	client 		*Client
	buff 		[]byte
	size 		int
	downloaded 	int
	requested 	int
	index		int
	block_size	int
}

func NewClient(infoHash, PeerID [20]byte, Peer decode.Peer) (*Client, error) {
	conn, err := net.DialTimeout("tcp", Peer.String(), 3 * time.Second);
	if err != nil {
		return nil, err;
	}
	_, err = HandShake(infoHash, PeerID, conn);
	if err != nil {
		conn.Close();
		return nil, err;
	}
	bt, err := RecvBT(conn)
	if err != nil {
		conn.Close();
		return nil, err;
	}
	return &Client {
		Choked: true,
		InfoHash: infoHash,
		PeerId: PeerID,
		Peer: Peer,
		conn: conn,
		bt_field: bt,
	}, nil;
}

func RecvBT(conn net.Conn) (bt.BT, error) {
	conn.SetDeadline(time.Now().Add(5 * time.Second));
	defer conn.SetDeadline(time.Time{});
	msg, err := ReadMSG(conn);
	if err != nil {
		return bt.BT{}, err;
	}
	if msg.ID != bitF {
		return bt.BT{}, fmt.Errorf("expected bitF, got=%T", msg.ID);
	}
	return msg.Payload, nil;
}

func HandShakeMSG(InfoHash, PeerId [20]byte) []byte {
	proto_name := "BitTorrent protocol";
	buff := make([]byte, len(proto_name) + 49);
	buff[0] = byte(len(proto_name));
	cur := 1
	cur += copy(buff[cur: ], proto_name);
	cur += copy(buff[cur: ], make([]byte, 8));
	cur += copy(buff[cur: ], InfoHash[:]);	
	cur += copy(buff[cur: ], PeerId[:]);
	return buff;
}

func Read(r io.Reader) ([]byte, error) {
	buff_len := make([]byte, 1);
	_, err := io.ReadFull(r, buff_len);
	if err != nil {
		return nil, err;
	}
	sz := int(buff_len[0]);
	if sz == 0 {
		return nil, fmt.Errorf("received 0 bytes");
	}
	buff := make([]byte, 48 + sz);
	_, err = io.ReadFull(r, buff);
	if err != nil {
		return nil, err;
	}
	return buff, nil;
}

func HandShake(InfoHash, PeerID [20]byte, conn net.Conn) ([]byte, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second));
	defer conn.SetDeadline(time.Time{});
	req := HandShakeMSG(InfoHash, PeerID);
	_ ,err := conn.Write(req);
	if err != nil {
		return nil, err;
	}
	buff, err := Read(conn);
	if err != nil {
		return nil, err;
	}
	sz := len("BitTorrent protocol");
	if !bytes.Equal(InfoHash[:], buff[sz + 8:sz + 28]) {
		return nil, fmt.Errorf("infohashes isn`t equal");
	}
	return buff, nil;
}

func ReadMSG(r io.Reader) (MSG, error) {
	buff_len := make([]byte, 4);
	_, err := io.ReadFull(r, buff_len);
	if err != nil {
		return MSG{}, err;
	}
	sz := binary.BigEndian.Uint32(buff_len[:]);
	if sz == 0 {
		return MSG{}, fmt.Errorf("received 0 bytes");
	}
	buff := make([]byte, sz);
	_, err = io.ReadFull(r, buff);
	if err != nil {
		return MSG{}, err;
	}
	return MSG {
		ID: int(buff[0]),
		Payload: buff[1:],
	}, nil;
}

func (m *MSG) Serialize() []byte {
	if m == nil {
		return make([]byte, 4);
	}
	sz := len(m.Payload) + 5;
	buff := make([]byte, sz);
	binary.BigEndian.PutUint32(buff[:4], uint32(sz-4));
	buff[4] = byte(m.ID);
	_ = copy(buff[5:], m.Payload);
	return buff;
}

func (c *Client) SendUnchoke() error {
	msg := MSG {
		ID: Unchoke,
	};
	// c.conn.SetDeadline(time.Now().Add(5 * time.Second));
	// defer c.conn.SetDeadline(time.Time{});
	sz, err := c.conn.Write(msg.Serialize());
	if sz != len(msg.Serialize()) {
		return fmt.Errorf("sent only %d bytes out of %d", sz, len(msg.Serialize()));
	}
	if err != nil {
		return err;
	}
	return nil;
}

func (c *Client) SendInterested() error {
	msg := MSG {
		ID: Interested,
	};
	// c.conn.SetDeadline(time.Now().Add(5 * time.Second));
	// defer c.conn.SetDeadline(time.Time{});
	sz, err := c.conn.Write(msg.Serialize());
	if sz != len(msg.Serialize()) {
		return fmt.Errorf("sent only %d bytes out of %d", sz, len(msg.Serialize()));
	}
	if err != nil {
		return err;
	}
	return nil;
}

func ParseHave(msg MSG) (int, error) {
	index := binary.BigEndian.Uint32(msg.Payload[1:]);
	if index < 0 {
		return -1, fmt.Errorf("received index < 0");
	}
	return int(index), nil;
}

func (p *ProgressInfo) ParsePiece(msg MSG, index int) (int, error) {
	data := msg.Payload;
	if len(data) < 8 {
		return -1, fmt.Errorf("msg payload too short");
	}
	recv_index := int(binary.BigEndian.Uint32(data[:4]));
	if index != recv_index {
		return -1, fmt.Errorf("index != received index");
	}
	begin := binary.BigEndian.Uint32(data[4:8]);
	data = data[8:];
	if p.size <= int(begin) {
		return -1, fmt.Errorf("begin >= size");
	}
	if len(data) + int(begin) > p.size {
		fmt.Println(int(begin) + len(data), " (<>) ", p.size);
		return -1, fmt.Errorf("len(data) + begin > size");
	}
	_ = copy(p.buff[begin:], data[:]);
	// fmt.Println("here!!!");
	return len(data), nil;
}


func (c *Client) Read() (MSG, error) {
	// c.conn.SetDeadline(time.Now().Add(5 * time.Second));
	// defer c.conn.SetDeadline(time.Time{});
	msg, err := ReadMSG(c.conn);
	if err != nil {
		return MSG{}, err;
	}
	return msg, nil;
}

func (p *ProgressInfo) Read() error {
	msg, err := p.client.Read();
	if err != nil {
		return err;
	}
	if msg.ID == Choke {
		p.client.Choked = true;
	} else if msg.ID == Unchoke {
		p.client.Choked = false;
	} else if msg.ID == Have {
		index, err := ParseHave(msg);
		if err != nil {
			return err;
		}
		p.client.bt_field.Set(index);
	} else if msg.ID == Pic {
		_, err := p.ParsePiece(msg, p.index);
		if err != nil {
			return err;
		}
		p.downloaded += p.block_size;
	}
	return nil;
}

func (c *Client) SendHave(index int) error {
	payload := make([]byte, 4);
	binary.BigEndian.PutUint32(payload[:], uint32(index));
	msg := MSG {
		ID: Have,
		Payload: payload[:],
	}
	_, err := c.conn.Write(msg.Serialize());
	if err != nil {
		return err;
	}
	return nil;
}

func (c *Client) sendRequest(index, begin, size int) error {
	payload := make([]byte, 12);
	binary.BigEndian.PutUint32(payload[:4], uint32(index));
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin));
	binary.BigEndian.PutUint32(payload[8:], uint32(size));
	msg := MSG {
		ID: Req,
		Payload: payload,
	};
	_, err := c.conn.Write(msg.Serialize());
	if err != nil {
		return err;
	}
	return nil;
}

func (c *Client) DownloadPiece(pic Piece) ([]byte, error) {
	var (
		size = pic.end - pic.begin
		block_size = 8192
	)
	p := ProgressInfo {
		client: c,
		size: size,
		downloaded: 0,
		requested: 0,
		block_size: block_size,
		index: pic.index,
		buff: make([]byte, size),
	};
	c.conn.SetDeadline(time.Now().Add(15 * time.Second));
	defer c.conn.SetDeadline(time.Time{});
	for ; p.downloaded < size; {
		if !c.Choked {
			if p.size - p.requested < block_size {
				p.block_size = p.size - p.requested;
			}
			err := c.sendRequest(pic.index, p.requested, p.block_size);
			if err != nil {
				return nil, err;
			}
			p.requested += p.block_size;
		}
		err := p.Read();
		if err != nil {
			return nil, err;
		}
		// fmt.Println(p.downloaded, "of", p.size);
	}
	
	return p.buff, nil;
}

// func Precalc(t decode.Torrent) Info { 
// 	...
// }

func Download(t decode.Torrent, index int) (Piece, error) {
	if index < 0 || index > len(t.PieceHashes) {
		return Piece{}, fmt.Errorf("index must be in range [0..%d]", len(t.PieceHashes));
	}
	pic := Piece {
		begin: index * t.PieceLength,
		end: index * t.PieceLength + t.PieceLength,
		hash: t.PieceHashes[index],
		index: index,
	};
	for _, peer := range t.Peers {
		c, err := NewClient(t.InfoHash, t.PeerID, peer);
		if err != nil {
			// return nil, err;
			fmt.Println(err);
			continue;
		}
		defer c.conn.Close();
		if !c.bt_field.Has(index) {
			// fmt.Println("continue...");
			continue;
		}
		c.SendUnchoke();
		c.SendInterested();
		buff, err := c.DownloadPiece(pic);
		if err != nil {
			fmt.Println(err);
			continue;
		}
		c.SendHave(pic.index);
		return Piece {
			buff: buff,
			index: pic.index,
		}, nil;
	}
	return Piece{}, fmt.Errorf("cannot download piece number %d because no one peer has it", index + 1);
}
