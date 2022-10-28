package app

import "net"

type Config struct {
	SourceAddress         string
	SourceEncryptKey      string
	DestinationAddress    string
	DestinationAddr       net.Addr
	DestinationEncryptKey string
}
