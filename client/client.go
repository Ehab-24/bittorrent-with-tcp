package client

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/Ehab-24/bittorrent/bitfield"
	"github.com/Ehab-24/bittorrent/handshake"
	"github.com/Ehab-24/bittorrent/message"
	"github.com/Ehab-24/bittorrent/peer"
)

type Client struct {
	Conn     net.Conn
	Choked   bool
	BitField bitfield.BitField
	infoHash [20]byte
	peerID   [20]byte
	peer     peer.Peer
}

func New(peer peer.Peer, peerID [20]byte, infoHash [20]byte) (*Client, error) {
	conn, err := net.DialTimeout("tcp", peer.AddrString(), 15*time.Second)
	if err != nil {
		return nil, err
	}

	if _, err := completeHandshake(conn, infoHash, peerID); err != nil {
		return nil, err
	}

	bf, err := recvBitField(conn)
	if err != nil {
		return nil, err
	}

	return &Client{
		Conn:     conn,
		Choked:   true,
		BitField: bf,
		peerID:   peerID,
		peer:     peer,
		infoHash: infoHash,
	}, nil
}

func completeHandshake(conn net.Conn, infoHash [20]byte, peerID [20]byte) (*handshake.HandShake, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{})

	req := handshake.New(infoHash, peerID)
	if _, err := conn.Write(req.Serialize()); err != nil {
		return nil, err
	}

	res, err := handshake.Read(conn)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(res.InfoHash[:], infoHash[:]) {
		return nil, fmt.Errorf("expected infoHash %v, but got %v", infoHash, res.InfoHash)
	}
	return res, nil
}

func recvBitField(conn net.Conn) (bitfield.BitField, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{})

	msg, err := message.Read(conn)
	if err != nil {
		return nil, err
	}
	if msg == nil || msg.ID != message.MsgBitfield {
		return nil, fmt.Errorf("expected bitField but got %v", msg)
	}
	return msg.Payload, nil
}

func (c *Client) Read() (*message.Message, error) {
	return message.Read(c.Conn)
}

func (c *Client) SendUnChoke() error {
	msg := message.Message{ID: message.MsgUnchoke}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendHave(index int) error {
	msg := message.FormatHave(index)
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendRequest(index int, begin int, length int) error {
	msg := message.FormatRequest(index, begin, length)
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendInterested() error {
	msg := message.Message{ID: message.MsgInterested}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendNotInterested() error {
	msg := message.Message{ID: message.MsgNotInterested}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}
