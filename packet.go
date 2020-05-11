package rudp

import (
	"encoding/binary"
	"errors"
)

// packet is a RUDP segment
type packet struct {
	seqNumber   uint32
	ackNumber   uint32
	segmentType rudpSegmentType
	payload     []byte
}

type rudpSegmentType int8

const (
	rudpSegmentTypeNormal rudpSegmentType = iota
	rudpSegmentTypeConn
	rudpSegmentTypeConnAck
	rudpSegmentTypeFin
	rudpSegmentTypeFinAck
	rudpSegmentTypeAck
	rudpSegmentTypePin
)

// ------------------------------------------
// |          Seq Number(32bits)            |
// ------------------------------------------
// |          Ack Number(32bits)            |
// ------------------------------------------
// |C|A|F|P|.15bits....|.......2byte........|
// |O|C|I|I|reserved...|.....data len.......|
// |N|K|N|N|...........|....................|
// -----------------------------------------
// |................data....................|

func unmarshal(data []byte) (*packet, error) {
	r := &packet{}
	if len(data) < RUDPHeaderLen {
		return nil, errors.New("illegal rudp segment")
	}

	r.seqNumber = binary.BigEndian.Uint32(data[0:4])
	r.ackNumber = binary.BigEndian.Uint32(data[4:8])

	connFlag, ackFlag, finFlag, pinFlag := parseFlag(data[8])
	switch {
	case connFlag:
		if ackFlag {
			r.segmentType = rudpSegmentTypeConnAck
		} else {
			r.segmentType = rudpSegmentTypeConn
		}
	case finFlag:
		if ackFlag {
			r.segmentType = rudpSegmentTypeFinAck
		} else {
			r.segmentType = rudpSegmentTypeFin
		}
	case pinFlag:
		r.segmentType = rudpSegmentTypePin
	case ackFlag:
		r.segmentType = rudpSegmentTypeAck
	default:
		r.segmentType = rudpSegmentTypeNormal
	}

	payloadLen := (binary.BigEndian.Uint16(data[10:12]))
	r.payload = data[RUDPHeaderLen : RUDPHeaderLen+payloadLen]
	return r, nil
}

func parseFlag(flag byte) (connFlag, ackFlag, finFlag, pinFlag bool) {
	if flag&(1<<7) != 0 {
		connFlag = true
	} else {
		connFlag = false
	}

	if flag&(1<<6) != 0 {
		ackFlag = true
	} else {
		ackFlag = false
	}

	if flag&(1<<5) != 0 {
		finFlag = true
	} else {
		finFlag = false
	}

	if flag&(1<<4) != 0 {
		pinFlag = true
	} else {
		pinFlag = false
	}
	return
}

func (p *packet) marshal() []byte {
	buf := make([]byte, RUDPHeaderLen+len(p.payload))
	binary.BigEndian.PutUint32(buf[0:4], p.seqNumber)
	binary.BigEndian.PutUint32(buf[4:8], p.ackNumber)
	var flag byte
	switch p.segmentType {
	case rudpSegmentTypeNormal:
		flag = 0
	case rudpSegmentTypeConn:
		flag &= (1 << 7)
	case rudpSegmentTypeConnAck:
		flag &= (1 << 7)
		flag &= (1 << 6)
	case rudpSegmentTypeFin:
		flag &= (1 << 5)
	case rudpSegmentTypeFinAck:
		flag &= (1 << 5)
		flag &= (1 << 4)
	case rudpSegmentTypePin:
		flag &= (1 << 3)
	}
	binary.BigEndian.PutUint16(buf[10:12], uint16(len(p.payload)))
	for i, b := range p.payload {
		buf[RUDPHeaderLen+i] = b
	}
	return buf
}
