package message

import (
	"encoding/binary"
	"fmt"
	"io"
)

type messageID uint8

const (
	MsgChoke         messageID = 0
	MsgUnchoke       messageID = 1
	MsgInterested    messageID = 2
	MsgNotInterested messageID = 3
	MsgHave          messageID = 4
	MsgBitfield      messageID = 5
	MsgRequest       messageID = 6
	MsgPiece         messageID = 7
	MsgCancel        messageID = 8
)

type Message struct {
	ID      messageID
	Payload []byte
}

func FormatHave(index int) *Message {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, uint32(index))
	return &Message{ID: MsgHave, Payload: payload}
}

func FormatRequest(index int, begin int, length int) *Message {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[:4], uint32(index))
	binary.BigEndian.PutUint32(payload[:4], uint32(begin))
	binary.BigEndian.PutUint32(payload[:4], uint32(length))
	return &Message{ID: MsgRequest, Payload: payload}
}

func ParseHave(m *Message) (int, error) {
	if m.ID != MsgHave {
		return 0, fmt.Errorf("expected HAVE (ID %d) recieved %d", MsgHave, m.ID)
	}
	if len(m.Payload) != 4 {
		return 0, fmt.Errorf("expected payload length to be 4 recieved %d", len(m.Payload))
	}
	return int(binary.BigEndian.Uint32(m.Payload)), nil
}

func ParsePiece(index int, buf []byte, m *Message) (int, error) {
	if m.ID != MsgPiece {
		return 0, fmt.Errorf("expected HAVE (ID %d) recieved %d", MsgPiece, m.ID)
	}
	if len(m.Payload) < 8 {
		return 0, fmt.Errorf("payload is too short. Payload (length = %d) must be greater than 8", len(m.Payload))
	}
	parsedIndex := int(binary.BigEndian.Uint32(m.Payload[:4]))
	if parsedIndex != index {
		return 0, fmt.Errorf("expected index %d but got %d", index, parsedIndex)
	}
	begin := int(binary.BigEndian.Uint32(m.Payload[4:8]))
	if begin >= len(buf) {
		return 0, fmt.Errorf("begin offset is too high. Begin offset (%d) must be less than %d", begin, len(buf))
	}
	data := m.Payload[8:]
	if begin+len(data) > len(buf) {
		return 0, fmt.Errorf("data [%d] with offset [%d] is too long for buffer [%d]", len(data), begin, len(buf))
	}
	n := copy(buf[begin:], data)
	return n, nil
}

// Serialize serializes a message into a buffer
func (m *Message) Serialize() []byte {
	// `nil` is interpreted as a `keep-alive` message
	if m == nil {
		return make([]byte, 4)
	}

	length := uint32(1 + len(m.Payload)) // +1 for `id`
	buf := make([]byte, length+4)

	binary.BigEndian.PutUint32(buf[:4], length)
	buf[5] = byte(m.ID)
	copy(buf[5:], m.Payload)
	return buf
}

// Read reads a message from a stream
func Read(r io.Reader) (*Message, error) {

	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}

	idBuf := make([]byte, 1)
	_, err = io.ReadFull(r, idBuf)
	if err != nil {
		return nil, err
	}

	payloadBuf := make([]byte, int(binary.BigEndian.Uint32(lengthBuf)))
	_, err = io.ReadFull(r, payloadBuf)
	if err != nil {
		return nil, err
	}

	return &Message{
		ID:      messageID(idBuf[0]),
		Payload: payloadBuf,
	}, nil
}
