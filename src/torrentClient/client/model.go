package client

import (
	"fmt"
	"net"

	"torrentClient/bitfield"
	"torrentClient/peers"
)

type Client struct {
	//Mu       sync.Mutex
	Conn     net.Conn
	Choked   bool
	Bitfield bitfield.Bitfield
	Peer     peers.Peer
	infoHash [20]byte
	peerID   [20]byte
}

func (c *Client) GetClientInfo() string {
	return fmt.Sprintf("Peer addr=%v\nChoked=%v\nBitfield: %v\n", c.Peer.GetAddr(), c.Choked, c.Bitfield)
}

func (c *Client) GetShortInfo() string {
	return fmt.Sprintf("Peer addr=%v, choked=%v", c.Peer.GetAddr(), c.Choked)
}

func (c *Client) GetPeer() peers.Peer {
	return c.Peer
}
