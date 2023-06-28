package handshake

import (
	"errors"
	"io"
	"log"
)

type HandShake struct {
	Psrt           string
	extensionBytes [8]byte
	InfoHash       [20]byte
	PeerID         [20]byte
}

func New(infoHash [20]byte, peerID [20]byte) *HandShake {
	return &HandShake{
		Psrt:           "BitTorrent protocol",
		extensionBytes: [8]byte{},
		PeerID:         peerID,
		InfoHash:       infoHash,
	}
}

// Serialize serializes the handshake to a buffer
func (h *HandShake) Serialize() []byte {
	buffer := make([]byte, 1+len(h.Psrt)+8+20+20)
	buffer[0] = byte(len(h.Psrt))

	index := 1
	index += copy(buffer[index:], []byte(h.Psrt))
	index += copy(buffer[index:], make([]byte, 8))
	index += copy(buffer[index:], h.InfoHash[:])
	index += copy(buffer[index:], h.PeerID[:])

	if index != len(buffer) {
		log.Println("Warning: Length of serialized buffer is invalid. Recheck values for all fields of the given buffer.")
	}

	return buffer
}

// Read parses a handshake from s stream
func Read(r io.Reader) (*HandShake, error) {

	lengthBuffer := make([]byte, 1)
	_, err := io.ReadFull(r, lengthBuffer)
	if err != nil {
		return nil, err
	}

	pstrLength := int(lengthBuffer[0])
	if pstrLength == 0 {
		return nil, errors.New("`pstr` length cannot be zero")
	}

	handShakeBuffer := make([]byte, pstrLength+8+20+20)
	_, err = io.ReadFull(r, handShakeBuffer)
	if err != nil {
		return nil, err
	}

	var extensionBytes [8]byte
	var infoHash, peerID [20]byte

	copy(extensionBytes[:], handShakeBuffer[pstrLength:pstrLength+8+20])
	copy(infoHash[:], handShakeBuffer[pstrLength:pstrLength+8+20])
	copy(peerID[:], handShakeBuffer[pstrLength+8+20:])

	return &HandShake{
		PeerID:         peerID,
		InfoHash:       infoHash,
		Psrt:           string(handShakeBuffer[:pstrLength]),
		extensionBytes: extensionBytes,
	}, nil
}
