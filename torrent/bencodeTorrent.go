package torrent

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"

	"github.com/jackpal/bencode-go"
)

type bencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type bencodeTrackerResponse struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

type bencodeTorrent struct {
	Announce string      `bencode:"announce" json:"announce,omitempty"`
	Info     bencodeInfo `bencode:"info" json:"info,omitempty"`
}

// Open parses the torrent file at the given path
func Open(path string) (TorrentFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return TorrentFile{}, err
	}
	defer file.Close()

	bt := bencodeTorrent{}
	if err = bencode.Unmarshal(file, &bt); err != nil {
		return TorrentFile{}, err
	}
	return bt.toTorrentFile()
}

// Returns the SHA-1 hash of the entire bencodeInfo
func (i *bencodeInfo) hash() ([sha1.Size]byte, error) {
	var buf bytes.Buffer
	if err := bencode.Marshal(&buf, *i); err != nil {
		return [sha1.Size]byte{}, err
	}
	return sha1.Sum(buf.Bytes()), nil
}

// Splits info `Pieces` (string) into and array of SHA-1 hash bytes
func (i *bencodeInfo) splitPieceHashes() ([][sha1.Size]byte, error) {
	buf := []byte(i.Pieces)
	if len(buf)%sha1.Size != 0 {
		return nil, fmt.Errorf("recieved malformed info. Length of `Pieces` [%d] should be a multiple of [%d]", len(buf), sha1.Size)
	}

	numHashes := len(buf) / sha1.Size
	hashes := make([][sha1.Size]byte, numHashes)
	for i := 0; i < numHashes; i++ {
		offset := i * sha1.Size
		copy(hashes[i][:], buf[offset:offset+sha1.Size])
	}
	return hashes, nil
}

func (bt *bencodeTorrent) toTorrentFile() (TorrentFile, error) {
	infoHash, err := bt.Info.hash()
	if err != nil {
		return TorrentFile{}, err
	}
	pieceHashes, err := bt.Info.splitPieceHashes()
	if err != nil {
		return TorrentFile{}, err
	}

	return TorrentFile{
		Announce:    bt.Announce,
		Name:        bt.Info.Name,
		PieceHashes: pieceHashes,
		PieceLength: bt.Info.PieceLength,
		InfoHash:    infoHash,
		Length:      bt.Info.Length,
	}, nil
}
