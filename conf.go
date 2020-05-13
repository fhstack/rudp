package rudp

import "fmt"

const (
	InitPeerMapCap              = 1 << 5
	RawUDPPacketLenLimit        = 576 - 8 - 60
	RUDPHeaderLen               = 12
	RUDPPPayloadLenLimit        = RawUDPPacketLenLimit - RUDPHeaderLen
	DefaultSendTickNano         = 1000 // 1ms
	DefaultHeartBeatCycleMinute = 30
	ResendDelayThreshholdMS     = 1000
)

const (
	RawUDPSendNotComplete = "raw udp not send the complete rudp packet"
)

var debug = false

func Debug() {
	debug = true
}

func log(format string, a ...interface{}) {
	if debug {
		fmt.Printf(format, a...)
	}
}