package rudp

const (
	InitPeerMapCap              = 1 << 5
	RawUDPPacketLenLimit        = 576 - 8 - 60
	RUDPHeaderLen               = 12
	RUDPPPayloadLenLimit        = RawUDPPacketLenLimit - RUDPHeaderLen
	DefaultSendTickNano         = 1000 // 1ms
	DefaultHeartBeatCycleMinute = 30
)

const (
	RawUDPSendNotComplete = "raw udp not send the complete rudp packet"
)
