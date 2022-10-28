package app

type Config struct {
	SourceAddress         string
	SourceEncryptKey      []byte
	DestinationAddress    string
	DestinationEncryptKey []byte
}
