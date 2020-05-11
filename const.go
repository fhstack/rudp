package rudp

const (
	InitPeerMapCap       = 1 << 5
	RawUDPPacketLenLimit = 576 - 8 - 60
	RUDPHeaderLen        = 12
)
