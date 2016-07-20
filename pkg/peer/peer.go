package peer

import "net"

// Peer A peer to connect to
type Peer struct {
	IP   net.IP
	Port uint16
}
