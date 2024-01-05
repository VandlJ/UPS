package structs

import "net"

type Player struct {
	Socket net.Conn
	Nick   string
}
