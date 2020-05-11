package rudp

import (
	"net"
	"time"
)

// RUDPConn is the reliable conn base on udp
type RUDPConn struct {
	sendPacketQueue   *packetList // all kinds of packet already to send
	resendPacketQueue *packetList // packet from sendPacketQueue that waiting for the peer's ack segment
	recvPacketQueue   *packetList // all kinds of packets recv

	recvData   chan []byte // Read Data
	localAddr  *net.UDPAddr
	remoteAddr *net.UDPAddr
	udpConn    *net.UDPConn

	closeConnCallBackListener func() // execute when connection closed
	buildConnCallBackListener func() // execute when connection build
}

// DialRUDP client dial server, building a relieable connection
func DialRUDP(localAddr *net.UDPAddr, remoteAddr *net.UDPAddr) (*RUDPConn, error) {
	c := &RUDPConn{}
	c.localAddr = localAddr
	c.remoteAddr = remoteAddr
	if err := c.clientDialBuildConn(); err != nil {
		return nil, err
	} else {
		return c, nil
	}

	// read and write loop
}

func (c *RUDPConn) clientDialBuildConn() error {
	// just init instance
	udpConn, err := net.DialUDP("udp", c.localAddr, c.remoteAddr)
	if err != nil {
		return err
	}
	// .....segment send and ack.......

	// succ
	c.udpConn = udpConn

	return nil
}

func NewRUDPConn() *RUDPConn {
	return &RUDPConn{}
}

func (c *RUDPConn) Read(b []byte) (int, error) {

	return 0, nil
}

func (c *RUDPConn) Write(b []byte) (int, error) {

	return 0, nil
}

func (c *RUDPConn) Close() error {

	return nil
}

func (c *RUDPConn) LocalAddr() net.Addr {

	return nil
}

func (c *RUDPConn) RemoteAddr() net.Addr {

	return nil
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
