package rudp

import (
	"errors"
	"math/rand"
	"net"
	"time"
)

const (
	maxWaitSegmentCntWhileConn = 3
)

var rander = rand.New(rand.NewSource(time.Now().Unix()))

type connStatus int8

const (
	connStatusConnecting connStatus = iota
	connStatusOpen
	connStatusClose
	connStatusErr
)

type connType int8

const (
	connTypeServer connType = iota
	connTypeClient
)

var (
	rudpConnClosedErr = errors.New("the rudp connection closed")
)

// RUDPConn is the reliable conn base on udp
type RUDPConn struct {
	sendPacketChannel chan *packet // all kinds of packet already to send
	resendPacketQueue *packetList  // packet from sendPacketQueue that waiting for the peer's ack segment

	recvPacketChannel chan *packet // all kinds of packets recv: ack/conn/fin/etc....
	outputPacketQueue *packetList

	localAddr  *net.UDPAddr
	remoteAddr *net.UDPAddr
	rawUDPConn *net.UDPConn

	sendSeqNumber uint32
	lastRecvTs    int64 // last recv data unix timestamp
	lastSendTs    int64

	sendTickNano        int32
	sendTickModifyEvent chan int32

	heartBeatCycleMinute int8

	closeConnCallback         func() // execute when connection closed
	buildConnCallbackListener func() // execute when connection build
	closeConnCallbackListener func()

	rudpConnStatus connStatus
	rudpConnType   connType

	sendStop      chan error
	resendStop    chan error
	recvStop      chan error
	heartbeatStop chan error

	err error
}

// DialRUDP client dial server, building a relieable connection
func DialRUDP(localAddr, remoteAddr *net.UDPAddr) (*RUDPConn, error) {
	c := &RUDPConn{}
	c.localAddr = localAddr
	c.remoteAddr = remoteAddr
	c.rudpConnType = connTypeClient
	c.sendSeqNumber = rander.Uint32()
	c.recvPacketChannel = make(chan *packet, 1<<5)
	c.rudpConnStatus = connStatusConnecting

	if err := c.clientBuildConn(); err != nil {
		return nil, err
	}

	c.sendPacketChannel = make(chan *packet, 1<<5)
	c.resendPacketQueue = newPacketList(packetListOrderBySeqNb)
	c.outputPacketQueue = newPacketList(packetListOrderBySeqNb)

	c.rudpConnStatus = connStatusOpen
	c.sendTickNano = DefaultSendTickNano
	c.sendTickModifyEvent = make(chan int32, 1)
	c.heartBeatCycleMinute = DefaultHeartBeatCycleMinute
	c.lastSendTs = time.Now().Unix()
	c.lastRecvTs = time.Now().Unix()

	c.sendStop = make(chan error, 1)
	c.recvStop = make(chan error, 1)
	c.resendStop = make(chan error, 1)
	c.heartbeatStop = make(chan error, 1)

	c.closeConnCallback = func() {
		c.rudpConnStatus = connStatusClose
		c.sendStop <- rudpConnClosedErr
		c.resendStop <- rudpConnClosedErr
		c.recvStop <- rudpConnClosedErr
		c.heartbeatStop <- rudpConnClosedErr
	}

	go c.recv()
	go c.send()
	go c.resend()
	// client need to keep a heart beat
	go c.keepLive()
	return c, nil
}

func (c *RUDPConn) keepLive() {
	tick := time.Tick(time.Duration(c.heartBeatCycleMinute) * 60 * time.Second)
	select {
	case <-tick:
		now := time.Now().Unix()
		if now-c.lastSendTs >= int64(c.heartBeatCycleMinute)*60 {
			c.sendPacketChannel <- newPinPacket()
		}
	case <-c.heartbeatStop:
		return
	}
}

func (c *RUDPConn) recv() {
	if c.rudpConnType == connTypeClient {
		c.clientRecv()
	} else if c.rudpConnType == connTypeServer {
		c.serverRecv()
	}
}

func (c *RUDPConn) clientRecv() {

}

func (c *RUDPConn) serverRecv() {

}

func (c *RUDPConn) send() {
	ticker := time.NewTicker(time.Duration(c.sendTickNano) * time.Nanosecond)
	for {
		select {
		case c.sendTickNano = <-c.sendTickModifyEvent:
			ticker.Stop()
			ticker = time.NewTicker(time.Duration(c.sendTickNano) * time.Nanosecond)
		case err := <-c.sendStop:
			c.err = err
			return
		default:
			select {
			case <-ticker.C:
				c.sendPacket()
				c.lastSendTs = time.Now().Unix()
			case err := <-c.sendStop:
				c.err = err
				return
			}

		}
	}
}

// SetRealSendTick modify the segment sending cycle
func (c *RUDPConn) SetSendTick(nano int32) {
	c.sendTickModifyEvent <- nano
}

func (c *RUDPConn) sendPacket() {
	apacket := <-c.sendPacketChannel
	segment := apacket.marshal()
	n, err := c.rawUDPConn.Write(segment)
	if err != nil {
		c.sendStop <- err
		return
	}
	if n != len(segment) {
		c.sendStop <- errors.New(RawUDPSendNotComplete)
		return
	}
	// only the normal segment possiblely needs to resend
	if apacket.segmentType == rudpSegmentTypeNormal {
		c.resendPacketQueue.putPacket(apacket)
	}
}

func (c *RUDPConn) clientBuildConn() error {
	// just init instance
	udpConn, err := net.DialUDP("udp", c.localAddr, c.remoteAddr)
	if err != nil {
		return err
	}
	c.rawUDPConn = udpConn

	// send conn segment
	connSeqNb := c.sendSeqNumber
	c.sendSeqNumber++
	connSegment := newConPacket(connSeqNb).marshal()
	n, err := udpConn.Write(connSegment)
	if err != nil {
		return err
	}
	if n != len(connSegment) {
		return errors.New(RawUDPSendNotComplete)
	}

	// wait the server ack conn segment
	// may the server's ack segment and normal segment out-of-order
	// so if the recv not the ack segment, we try to wait the next
	for cnt := 0; cnt < maxWaitSegmentCntWhileConn; cnt++ {
		buf := make([]byte, RawUDPPacketLenLimit)
		n, err = udpConn.Read(buf)
		if err != nil {
			return err
		}
		recvPacket, err := unmarshalRUDPPacket(buf)
		if err != nil {
			return errors.New("analyze the recvSegment error: " + err.Error())
		}

		if recvPacket.ackNumber == connSeqNb && recvPacket.segmentType == rudpSegmentTypeConnAck {
			// conn OK
			return nil
		} else {
			c.recvPacketChannel <- recvPacket
			continue
		}
	}
	return nil
}

func serverBuildConn(localAddr, remoteAddr *net.UDPAddr) (*RUDPConn, error) {
	c := &RUDPConn{}
	udpConn, err := net.DialUDP("udp", c.localAddr, c.remoteAddr)
	if err != nil {
		return nil, err
	}
	c.rawUDPConn = udpConn

	c.localAddr = localAddr
	c.remoteAddr = remoteAddr
	c.rudpConnType = connTypeServer
	c.sendSeqNumber = rander.Uint32()
	c.recvPacketChannel = make(chan *packet, 1<<5)
	c.rudpConnStatus = connStatusConnecting

	c.sendPacketChannel = make(chan *packet, 1<<5)
	c.resendPacketQueue = newPacketList(packetListOrderBySeqNb)
	c.outputPacketQueue = newPacketList(packetListOrderBySeqNb)

	c.sendTickNano = DefaultSendTickNano
	c.sendTickModifyEvent = make(chan int32, 1)
	c.lastRecvTs = time.Now().Unix()

	c.sendStop = make(chan error, 1)
	c.recvStop = make(chan error, 1)

	c.closeConnCallback = func() {
		c.rudpConnStatus = connStatusClose
		closed := errors.New("the rudp connection closed")
		c.sendStop <- closed
		c.recvStop <- closed
		c.heartbeatStop <- closed
	}

	go c.send()
	go c.recv()
	go c.resend()

	return c, nil
}

func (c *RUDPConn) Read(b []byte) (int, error) {

	return 0, nil
}

func (c *RUDPConn) Write(b []byte) (int, error) {
	if c.err != nil {
		return 0, errors.New("rudp write error: " + c.err.Error())
	}
	n := len(b)
	for {
		if len(b) <= RUDPPPayloadLenLimit {
			c.sendPacketChannel <- newNormalPacket(b, c.sendSeqNumber)
			c.sendSeqNumber++
			return n, nil
		} else {
			c.sendPacketChannel <- newNormalPacket(b[:RUDPPPayloadLenLimit], c.sendSeqNumber)
			c.sendSeqNumber++
			b = b[RUDPPPayloadLenLimit:]
		}
	}
}

// Close close must be called while closing the conn
func (c *RUDPConn) Close() error {
	if c.rudpConnStatus != connStatusOpen {
		return errors.New("the rudp connection is not open status!")
	}
	defer func() {
		if c.rudpConnType == connTypeServer {
			c.closeConnCallbackListener()
		}
		c.closeConnCallback()
	}()

	finSegment := newFinPacket().marshal()
	n, err := c.rawUDPConn.Write(finSegment)
	if err != nil {
		return err
	}
	if n != len(finSegment) {
		return errors.New(RawUDPSendNotComplete)
	}
	return nil
}

func (c *RUDPConn) resend() {

}

func (c *RUDPConn) LocalAddr() net.Addr {
	return c.localAddr
}

func (c *RUDPConn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *RUDPConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *RUDPConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *RUDPConn) SetWriteDeadline(t time.Time) error {
	return nil
}
