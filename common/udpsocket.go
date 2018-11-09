package common

import (
    "net"
)

// A UDPSocket is an abstraction of a typical UDP socket and provides
// higher-level functions to create UDP connections.
type UDPSocket struct {
    connection *net.UDPConn
}

// Create a new UDP socket and bind it to the given port.
func NewUDPSocket(address string) *UDPSocket {

    udpAddr, err := net.ResolveUDPAddr("udp4", address)
    if err != nil { panic(err) }

    udpConn, err := net.ListenUDP("udp4", udpAddr)
    if err != nil { panic(err) }

    return &UDPSocket{udpConn}
}

// Wait until new data is receives and extract it. Also return
// the address from which the data comes and a flag indicating
// whether the connection is still alive or not.
func (socket *UDPSocket) Receive() ([]byte, string, bool) {

    buffer := make([]byte, SocketBufferSize)

    n, peer, err := socket.connection.ReadFromUDP(buffer)
    if err != nil { return []byte{}, "", false }

    return buffer[:n], peer.String(), true
}

// Send data to another address using the socket UDP connection
func (socket *UDPSocket) Send(bytes []byte, address string) {

    udpAddr, err := net.ResolveUDPAddr("udp4", address)
    if err != nil { panic(err) }

    _, err = socket.connection.WriteToUDP(bytes, udpAddr)
    if err != nil { panic(err)  }

}

// Close the connection
func (socket *UDPSocket) Unbind() {
    socket.connection.Close()
}