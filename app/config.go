package app

type Config struct {
	LogLevel              string
	SourceAddress         string
	SourceEncryptKey      []byte
	DestinationAddress    string
	DestinationEncryptKey []byte
}
