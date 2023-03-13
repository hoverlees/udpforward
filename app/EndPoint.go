package app

import (
	"strings"
	"udpforward/logger"
)

type endpointChannelMsg struct {
	connId string
	data   []byte
}

type SourceEndpoint interface {
	Init(address string, encryptKey []byte)
	Listen()
	ReadPacket() (connId string, data []byte)
	WritePacket(connId string, data []byte)
}

type DestinationEndpoint interface {
	Init(address string, encryptKey []byte)
	ReadPacket() (connId string, data []byte)
	WritePacket(connId string, data []byte)
}

func xorData(src []byte, key []byte) []byte {
	data := make([]byte, len(src))
	copy(data, src)
	if len(key) == 0 {
		return data
	}
	dataLen := len(data)
	keyLen := len(key)
	keyIndex := 0
	for i := 0; i < dataLen; i++ {
		data[i] = data[i] ^ key[keyIndex]
		keyIndex++
		if keyIndex >= keyLen {

			keyIndex = 0
		}
	}
	return data
}

func NewSourceEndpoint(address string, encryptKey []byte) SourceEndpoint {
	if strings.Index(address, "tcp://") == 0 {
		endpoint := &TcpSourceEndpoint{}
		endpoint.Init(address[6:], encryptKey)
		return endpoint
	} else if strings.Index(address, "udp://") == 0 {
		endpoint := &UdpSourceEndpoint{}
		endpoint.Init(address[6:], encryptKey)
		return endpoint
	} else {
		logger.Panic("cant create source endpoint for address %s, tcp:// or udp:// is needed.", address)
	}
	return nil
}

func NewDestinationEndpoint(address string, encryptKey []byte) DestinationEndpoint {
	if strings.Index(address, "tcp://") == 0 {
		endpoint := &TcpDestinationEndpoint{}
		endpoint.Init(address[6:], encryptKey)
		return endpoint
	} else if strings.Index(address, "udp://") == 0 {
		endpoint := &UdpDestinationEndpoint{}
		endpoint.Init(address[6:], encryptKey)
		return endpoint
	} else {
		logger.Panic("cant create destination endpoint for address %s, tcp:// or udp:// is needed.", address)
	}
	return nil
}
