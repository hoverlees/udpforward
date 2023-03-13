package app

import (
	"net"
	"sync"
	"time"
	"udpforward/logger"
)

type UdpSourceEndpoint struct {
	SourceEndpoint
	address     string
	encryptKey  []byte
	conn        net.PacketConn
	clientCache *Cache
	dataChannel chan endpointChannelMsg
}

func (e *UdpSourceEndpoint) Init(address string, encryptKey []byte) {
	e.address = address
	e.encryptKey = encryptKey
	e.dataChannel = make(chan endpointChannelMsg, 1024)
	e.clientCache = NewCache(1800)
}

func (e *UdpSourceEndpoint) Listen() {
	conn, err := net.ListenPacket("udp", e.address)
	if err != nil {
		logger.Error("udp source endpoint listen fail, %s", err)
		return
	}
	e.conn = conn
	go e.loopRead()
}

func (e *UdpSourceEndpoint) loopRead() {
	buffer := make([]byte, 10240)
	for {
		n, addr, err := e.conn.ReadFrom(buffer)
		if err != nil {
			logger.Debug("read udp fail, %s", err)
			break
		}
		connId := addr.String()
		e.clientCache.Add(connId, addr)
		data := xorData(buffer[0:n], e.encryptKey)
		logger.Debug("udp source endpoint received %d bytes, addr=%v, data=%v", n, connId, data)
		msg := endpointChannelMsg{
			connId: connId,
			data:   data,
		}
		e.dataChannel <- msg
	}
}

func (e *UdpSourceEndpoint) ReadPacket() (connId string, data []byte) {
	msg := <-e.dataChannel
	return msg.connId, msg.data
}

func (e *UdpSourceEndpoint) WritePacket(connId string, data []byte) {
	v := e.clientCache.Get(connId)
	if v == nil {
		logger.Debug("udp conn %s closed", connId)
		return
	}
	addr := v.(net.Addr)
	data = xorData(data, e.encryptKey)
	logger.Debug("udp source endpoint write %d bytes, addr=%v", len(data), addr)
	e.conn.WriteTo(data, addr)
}

type UdpDestinationEndpoint struct {
	DestinationEndpoint
	address         string
	encryptKey      []byte
	destinationAddr net.Addr
	connMap         *sync.Map
	dataChannel     chan endpointChannelMsg
}

func (e *UdpDestinationEndpoint) Init(address string, encryptKey []byte) {
	e.address = address
	e.encryptKey = encryptKey
	destinationAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		logger.Panic("can't resolv udp address %s", address)
		return
	}
	e.destinationAddr = destinationAddr
	e.connMap = &sync.Map{}
	e.dataChannel = make(chan endpointChannelMsg, 1024)
}

func (e *UdpDestinationEndpoint) ReadPacket() (connId string, data []byte) {
	msg := <-e.dataChannel
	return msg.connId, msg.data
}

func (e *UdpDestinationEndpoint) WritePacket(connId string, data []byte) {
	var conn net.PacketConn
	v, ok := e.connMap.Load(connId)
	if !ok {
		c, err := net.ListenPacket("udp", ":0")
		if err != nil {
			logger.Debug("fail to listen udp for %s", connId)
			return
		}
		e.connMap.Store(connId, c)
		go e.startReadConn(c, connId)
		conn = c
	} else {
		conn = v.(net.PacketConn)
	}
	logger.Debug("udp destination endpoint write %d bytes, addr=%v", len(data), connId)
	data = xorData(data, e.encryptKey)
	conn.WriteTo(data, e.destinationAddr)
}

func (e *UdpDestinationEndpoint) startReadConn(conn net.PacketConn, connId string) {
	buffer := make([]byte, 10240)
	for {
		conn.SetReadDeadline(time.Now().Add(time.Hour))
		n, _, err := conn.ReadFrom(buffer)
		if err != nil {
			logger.Debug("read udp error, %s", err)
			break
		}
		logger.Debug("udp destination endpoint read %d bytes, addr=%v", n, connId)
		data := xorData(buffer[0:n], e.encryptKey)
		msg := endpointChannelMsg{
			connId: connId,
			data:   data,
		}
		e.dataChannel <- msg
	}
	e.connMap.Delete(connId)
}
