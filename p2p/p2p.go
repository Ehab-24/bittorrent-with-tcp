package p2p

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"time"

	"github.com/Ehab-24/bittorrent/client"
	"github.com/Ehab-24/bittorrent/message"
	"github.com/Ehab-24/bittorrent/peer"
)

const (
	MaxBacklog = 5

	MaxBlockSize = 16 * 1024 // 16kB
)

type Torrent struct {
	Peers       []peer.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type pieceResult struct {
	index int
	buf   []byte
}

type pieceProgress struct {
	index      int
	client     *client.Client
	buf        []byte
	downloaded int
	requested  int
	backlog    int
}

func (state *pieceProgress) readMessage() error {
	msg, err := state.client.Read()
	switch msg.ID {
	case message.MsgChoke:
		state.client.Choked = true
	case message.MsgUnchoke:
		state.client.Choked = false
	case message.MsgHave:
		var index int
		index, err = message.ParseHave(msg)
		state.client.BitField.SetPiece(index)
	case message.MsgPiece:
		var n int
		n, err = message.ParsePiece(state.index, state.buf, msg)
		state.downloaded += n
		state.backlog--
	}
	return err
}

// Returns `t.PieceLength` for every piece except the last one.
func (t *Torrent) calculatePieceSize(index int) int {
	begin, end := t.calculateBoundsForPiece(index)
	return end - begin
}

func (t *Torrent) calculateBoundsForPiece(index int) (int, int) {
	begin := index * t.PieceLength
	end := begin + t.PieceLength
	if end > t.Length {
		end = t.Length
	}
	return begin, end
}

func checkIntegrity(pw *pieceWork, buf []byte) error {
	bufHash := sha1.Sum(buf)
	if !bytes.Equal(pw.hash[:], bufHash[:]) {
		return fmt.Errorf("! Piece %v failed integrity check", pw.index)
	}
	return nil
}

func downloadPiece(c *client.Client, pw *pieceWork) ([]byte, error) {
	state := pieceProgress{
		client: c,
		index:  pw.index,
		buf:    make([]byte, pw.length),
	}

	c.Conn.SetDeadline(time.Now().Add(15 * time.Second))
	defer c.Conn.SetDeadline(time.Time{})

	for state.downloaded < pw.length {
		if !state.client.Choked {
			for state.backlog < MaxBacklog && state.requested < pw.length {
				blockSize := MaxBlockSize

				// Last block might be shorter than the typical block
				if pw.length-state.requested < blockSize {
					blockSize = pw.length - state.requested
				}

				if err := c.SendRequest(pw.index, state.requested, blockSize); err != nil {
					return nil, err
				}
				state.backlog++
				state.requested += blockSize
			}
		}

		if err := state.readMessage(); err != nil {
			return nil, err
		}
	}

	return state.buf, nil
}

func (t *Torrent) startDownloadWorker(peer peer.Peer, workQueue chan *pieceWork, results chan *pieceResult) {
	client, err := client.New(peer, t.PeerID, t.InfoHash)
	defer client.Conn.Close()
	if err != nil {
		log.Println("Unable to establish handshake with", peer.IP)
		return
	}

	client.SendUnChoke()
	client.SendInterested()

	for pieceWork := range workQueue {
		if !client.BitField.HasPiece(pieceWork.index) {
			workQueue <- pieceWork
			continue
		}

		buf, err := downloadPiece(client, pieceWork)
		if err != nil {
			log.Println("Error: ", err)
			workQueue <- pieceWork
			return
		}

		if err = checkIntegrity(pieceWork, buf); err != nil {
			log.Printf("Piece %v failed integrity check", pieceWork.index)
			workQueue <- pieceWork
			continue
		}

		client.SendHave(pieceWork.index)
		results <- &pieceResult{index: pieceWork.index, buf: buf}
	}
}

func (t *Torrent) Download() ([]byte, error) {
	log.Printf("Starting %v for download", t.Name)

	results := make(chan *pieceResult)
	workQueue := make(chan *pieceWork, len(t.PieceHashes))
	defer close(workQueue)

	for i, hash := range t.PieceHashes {
		length := t.calculatePieceSize(i) // Size (in bytes) of the current file piece
		workQueue <- &pieceWork{index: i, length: length, hash: hash}
	}

	for _, peer := range t.Peers {
		go t.startDownloadWorker(peer, workQueue, results)
	}

	donePieces := 0
	fileBuf := make([]byte, t.Length)
	// ! Currently, the whole is stored as a buffer in memory
	// TODO: save to disk
	for donePieces < len(t.PieceHashes) {
		res := <-results
		begin, end := t.calculateBoundsForPiece(res.index) // Calculate where the pice should sit in the file buffer
		copy(fileBuf[begin:end], res.buf)
		donePieces++
	}

	return fileBuf, nil
}
