package main

import (
	"flag"
	"log"
	"net"
	"os"
	"sync"
	"time"
	"udpforward/app"
)

// udpforward --source-address=:9999 --source-key=adfadsfasdfadsf --destination-address=hoverlees.com:3600 --destination-key=123412341234
var appConfig *app.Config = &app.Config{}
var targetConnMap *sync.Map = &sync.Map{}

func startTunnelSession(readConn net.PacketConn, toConn net.PacketConn, toAddr net.Addr) {
	for {
		readConn.SetReadDeadline(time.Now().Add(time.Minute * 15))
		buffer := make([]byte, 2048)
		n, addr, err := readConn.ReadFrom(buffer)
		if err != nil {
			log.Print(err)
			break
		}
		log.Printf("received %d bytes, from addr=%v", n, addr)
		toConn.WriteTo(buffer[0:n], toAddr)
	}
	targetConnMap.Delete(toAddr.String())
}

func main() {
	flag.StringVar(&appConfig.SourceAddress, "source-address", "", "The source address to listen and receive source data")
	flag.StringVar(&appConfig.SourceEncryptKey, "source-encrypt-key", "", "The source encrypt key for data decription, leave empty means do not decription source data")
	flag.StringVar(&appConfig.DestinationAddress, "destination-address", "", "The destination address to send source data to")
	flag.StringVar(&appConfig.DestinationEncryptKey, "destination-encrypt-key", "", "The destination encrypt key for data encryption, leave empty means do not encrypt data send to destination")
	flag.Parse()
	if appConfig.DestinationAddress == "" || appConfig.SourceAddress == "" {
		log.Fatal("source address or destination address must be set.")
		os.Exit(1)
	}
	destinationAddr, err := net.ResolveUDPAddr("udp", appConfig.DestinationAddress)
	if err != nil {
		log.Fatalf("can't parse destination address %s, err=%s", appConfig.DestinationAddress, err)
	}
	appConfig.DestinationAddr = destinationAddr

	conn, err := net.ListenPacket("udp", appConfig.SourceAddress)
	if err != nil {
		log.Fatal(err)
	}
	buffer := make([]byte, 2048)
	for {
		n, addr, err := conn.ReadFrom(buffer)
		if err != nil {
			log.Print(err)
			break
		}
		log.Printf("received %d bytes, addr=%v", n, addr)
		var targetConn net.PacketConn
		tConn, ok := targetConnMap.Load(addr.String())
		if !ok {
			targetConn, err = net.ListenPacket("udp", ":0")
			if err != nil {
				log.Printf("fail to listen udp for %s", addr.String())
				continue
			}
			targetConnMap.Store(addr.String(), targetConn)
			go startTunnelSession(targetConn, conn, addr)
		} else {
			targetConn = tConn.(net.PacketConn)
		}
		if len(appConfig.SourceEncryptKey) > 0 {
			//TODO decrypt
		}
		targetConn.WriteTo(buffer[0:n], appConfig.DestinationAddr)
	}
}
