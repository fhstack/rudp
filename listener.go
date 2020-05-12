package rudp

import (
	"errors"
	"net"
	"sync"
)

// RUDPListener implements net.Listener interface
type RUDPListener struct {
	listenConn   *net.UDPConn
	peerMap      *sync.Map
	newConnQueue chan *RUDPConn
	err          chan error
}

func ListenRUDP(localAddr *net.UDPAddr) (*RUDPListener, error) {
	udpConn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return nil, err
	}

	listener := &RUDPListener{
		listenConn:   udpConn,
		peerMap:      &sync.Map{},
		err:          make(chan error, 1),
		newConnQueue: make(chan *RUDPConn, 1<<10),
	}

	return listener, nil
}

func (l *RUDPListener) Accept() (net.Conn, error) {
	select {
	case c, ok := <-l.newConnQueue:
		if !ok {
			return nil, errors.New("listener closed")
		}
		return c, nil
	case err := <-l.err:
		return nil, err
	}
	return nil, nil
}

func (l *RUDPListener) Close() error {

	return nil
}

func (l *RUDPListener) Addr() net.Addr {
	return l.listenConn.LocalAddr()
}

func (l *RUDPListener) listen() {
	for {
		buf := make([]byte, RawUDPPacketLenLimit)
		n, remoteAddr, err := l.listenConn.ReadFromUDP(buf)
		if err != nil {
			l.err <- err
			return
		}
		buf = buf[:n]
		if v, ok := l.peerMap.Load(remoteAddr.String()); ok {
			rudpConn := v.(*RUDPConn)
			apacket, err := unmarshalRUDPPacket(buf)
			if err != nil {
				l.err <- err
				return
			}
			rudpConn.recvPacketChannel <- apacket
		} else {
			localAddr, _ := net.ResolveUDPAddr(l.Addr().Network(), l.Addr().String())
			rudpConn, err := serverBuildConn(localAddr, remoteAddr)
			if err != nil {
				l.err <- err
				return
			}
			rudpConn.buildConnCallbackListener = func() {
				l.newConnQueue <- rudpConn
			}
			rudpConn.closeConnCallbackListener = func() {
				l.peerMap.Delete(rudpConn.remoteAddr)
			}
			l.peerMap.Store(remoteAddr, rudpConn)
		}
	}
}
