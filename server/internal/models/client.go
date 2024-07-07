package models

import (
	"net"
	"time"
)

type Client struct {
	Addr       *net.UDPAddr
	Username   string
	LastActive time.Time
}
