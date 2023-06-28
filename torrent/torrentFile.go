package torrent

import (
	"crypto/rand"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/Ehab-24/bittorrent/p2p"
	"github.com/Ehab-24/bittorrent/peer"
	"github.com/jackpal/bencode-go"
)

const Port uint16 = 6881

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

// func (bto *bencodeTorrent) toTorrentFile() (TorrentFile, error) {

// 	hash, err :=
// }

// DownloadToFile downloads the TorrentFile and places content at the given path
func (t *TorrentFile) DownloadToFile(path string) error {
	var peerID [20]byte
	if _, err := rand.Read(peerID[:]); err != nil {
		return err
	}
	peers, err := t.requestPeers(peerID, Port)
	if err != nil {
		return err
	}

	log.Println("good 2")

	torrent := p2p.Torrent{
		Peers:       peers,
		PeerID:      peerID,
		InfoHash:    t.InfoHash,
		PieceHashes: t.PieceHashes,
		PieceLength: t.PieceLength,
		Length:      t.Length,
		Name:        t.Name,
	}
	log.Println("good 3")

	buf, err := torrent.Download()
	if err != nil {
		return err
	}

	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = dst.Write(buf)
	return err
}

func (t *TorrentFile) buildTrackerURL(peerID [20]byte, port uint16) (string, error) {
	base, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}

	queryParams := url.Values{
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(int(port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(t.Length)},
	}
	base.RawQuery = queryParams.Encode()

	log.Println(t.Announce)

	return base.String(), nil
}

func (t *TorrentFile) requestPeers(peerID [20]byte, port uint16) ([]peer.Peer, error) {
	url, err := t.buildTrackerURL(peerID, port)
	if err != nil {
		return nil, err
	}

	c := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := c.Get(url)
	if err != nil {
		log.Println("found it")
		return nil, err
	}
	defer resp.Body.Close()

	trackerResp := bencodeTrackerResponse{}
	if err = bencode.Unmarshal(resp.Body, &trackerResp); err != nil {
		return nil, err
	}

	return peer.Unmarshal([]byte(trackerResp.Peers))
}
