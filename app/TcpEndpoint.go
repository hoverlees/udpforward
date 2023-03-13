package app

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"udpforward/logger"
)

type tcpDataParser struct {
	onPacketData  func(data []byte)
	parseStep     int //0-size0, 1-size1, 2-data
	packetSize    int
	packetReadPos int
	packetBuffer  []byte
}

func newTcpDataParser() *tcpDataParser {
	parser := &tcpDataParser{
		parseStep:    0,
		packetBuffer: make([]byte, 10240),
	}
	return parser
}

func (p *tcpDataParser) addByte(b byte) {
	if p.parseStep == 2 {
		p.packetBuffer[p.packetReadPos] = b
		p.packetReadPos++
		if p.packetReadPos >= p.packetSize {
			p.onPacketData(p.packetBuffer[0:p.packetSize])
			p.parseStep = 0
		}
	} else if p.parseStep == 1 {
		p.packetSize = p.packetSize + (int(b) << 8)
		p.parseStep = 2
		p.packetReadPos = 0
	} else if p.parseStep == 0 {
		p.packetSize = int(b)
		p.parseStep = 1
	}
}

type TcpSourceEndpoint struct {
	SourceEndpoint
	address     string
	encryptKey  []byte
	connIdGen   uint64
	connMap     *sync.Map
	dataChannel chan endpointChannelMsg
}

type TcpDestinationEndpoint struct {
	DestinationEndpoint
	address     string
	encryptKey  []byte
	connMap     *sync.Map
	dataChannel chan endpointChannelMsg
}

func (e *TcpSourceEndpoint) Init(address string, encryptKey []byte) {
	e.address = address
	e.encryptKey = encryptKey
	e.connMap = &sync.Map{}
	e.connIdGen = 1
	e.dataChannel = make(chan endpointChannelMsg, 1024)
}

func (e *TcpSourceEndpoint) Listen() {
	go func() {
		logger.Debug("start listen tcp %s", e.address)
		ln, err := net.Listen("tcp", e.address)
		if err != nil {
			logger.Panic("can't listen source endpoint for address %s", e.address)
			return
		}
		for {
			conn, err := ln.Accept()
			if err != nil {
				logger.Debug("source endpoint accept failure, %s", err)
				continue
			}
			go e.handleClientConnection(conn)
		}
	}()
}

func (e *TcpSourceEndpoint) handleClientConnection(conn net.Conn) {
	connId := fmt.Sprintf("%d", atomic.AddUint64(&e.connIdGen, 1))
	e.connMap.Store(connId, conn)
	logger.Debug("new tcp source connect %s connect", connId)
	readBuffer := make([]byte, 1024)
	parser := newTcpDataParser()
	parser.onPacketData = func(data []byte) {
		data = xorData(data, e.encryptKey)
		msg := endpointChannelMsg{
			connId: connId,
			data:   data,
		}
		e.dataChannel <- msg
	}
	defer conn.Close()
	for {
		conn.SetDeadline(time.Now().Add(time.Minute))
		n, err := conn.Read(readBuffer)
		if err != nil {
			e.connMap.Delete(connId)
			return
		}
		for i := 0; i < n; i++ {
			parser.addByte(readBuffer[i])
		}
	}
}

func (e *TcpSourceEndpoint) ReadPacket() (connId string, data []byte) {
	msg := <-e.dataChannel
	return msg.connId, msg.data
}

func (e *TcpSourceEndpoint) WritePacket(connId string, data []byte) {
	v, ok := e.connMap.Load(connId)
	if !ok {
		logger.Debug("source endpoint can't find conn %s to write packet", connId)
		return
	}
	conn := v.(net.Conn)
	sizeByte := make([]byte, 2)
	data = xorData(data, e.encryptKey)
	size := len(data)
	sizeByte[0] = byte(size & 0xff)
	sizeByte[1] = byte((size >> 8) & 0xff)
	conn.Write(sizeByte)
	conn.Write(data)
}

//======

func (e *TcpDestinationEndpoint) Init(address string, encryptKey []byte) {
	e.address = address
	e.encryptKey = encryptKey
	e.connMap = &sync.Map{}
	e.dataChannel = make(chan endpointChannelMsg, 1024)
}

func (e *TcpDestinationEndpoint) ReadPacket() (connId string, data []byte) {
	msg := <-e.dataChannel
	return msg.connId, msg.data
}

func (e *TcpDestinationEndpoint) handleClientConnection(connId string, conn net.Conn) {
	readBuffer := make([]byte, 1024)
	parser := newTcpDataParser()
	parser.onPacketData = func(data []byte) {
		data = xorData(data, e.encryptKey)
		msg := endpointChannelMsg{
			connId: connId,
			data:   data,
		}
		e.dataChannel <- msg
	}
	defer conn.Close()
	for {
		conn.SetDeadline(time.Now().Add(time.Minute))
		n, err := conn.Read(readBuffer)
		if err != nil {
			e.connMap.Delete(connId)
			return
		}
		for i := 0; i < n; i++ {
			parser.addByte(readBuffer[i])
		}
	}
}

func (e *TcpDestinationEndpoint) WritePacket(connId string, data []byte) {
	v, ok := e.connMap.Load(connId)
	if !ok {
		c, err := net.Dial("tcp", e.address)
		if err != nil {
			logger.Info("can't dail to %s", e.address)
			return
		}
		e.connMap.Store(connId, c)
		go e.handleClientConnection(connId, c)
		v = c
	}
	conn := v.(net.Conn)
	sizeByte := make([]byte, 2)
	data = xorData(data, e.encryptKey)
	size := len(data)
	sizeByte[0] = byte(size & 0xff)
	sizeByte[1] = byte((size >> 8) & 0xff)
	conn.Write(sizeByte)
	conn.Write(data)
}
