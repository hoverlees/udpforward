package main

import (
	"flag"
	"os"
	"udpforward/app"
	"udpforward/logger"
)

var appConfig *app.Config = &app.Config{}

func main() {
	sourceEncryptKey := ""
	destinationEncryptKey := ""
	flag.StringVar(&appConfig.LogLevel, "log-level", "INFO", "The log level, can be DEBUG, INFO, WARN, ERROR, default is INFO")
	flag.StringVar(&appConfig.SourceAddress, "source-address", "", "The source address to listen and receive source data. The format is tcp://ip:port or udp://ip:port")
	flag.StringVar(&sourceEncryptKey, "source-encrypt-key", "", "The source encrypt key for data decription, leave empty means do not decription source data")
	flag.StringVar(&appConfig.DestinationAddress, "destination-address", "", "The destination address to send source data to. The format is tcp://ip:port or udp://ip:port")
	flag.StringVar(&destinationEncryptKey, "destination-encrypt-key", "", "The destination encrypt key for data encryption, leave empty means do not encrypt data send to destination")
	flag.Parse()
	logger.SetLogLevelByName(appConfig.LogLevel)
	if appConfig.DestinationAddress == "" || appConfig.SourceAddress == "" {
		logger.Panic("source address or destination address must be set.")
		os.Exit(1)
	}
	appConfig.SourceEncryptKey = []byte(sourceEncryptKey)
	appConfig.DestinationEncryptKey = []byte(destinationEncryptKey)
	sourceEndpoint := app.NewSourceEndpoint(appConfig.SourceAddress, appConfig.SourceEncryptKey)
	destinationEndpoint := app.NewDestinationEndpoint(appConfig.DestinationAddress, appConfig.DestinationEncryptKey)
	sourceEndpoint.Listen()
	logger.Debug("udp forward started, from %s to %s", appConfig.SourceAddress, appConfig.DestinationAddress)
	go func() {
		//forward destination to source
		for {
			connId, data := destinationEndpoint.ReadPacket()
			logger.Debug("destination receive data: %v", data)
			sourceEndpoint.WritePacket(connId, data)
		}
	}()
	for {
		connId, data := sourceEndpoint.ReadPacket()
		logger.Debug("source recieve data: %v", data)
		destinationEndpoint.WritePacket(connId, data)
	}
}
