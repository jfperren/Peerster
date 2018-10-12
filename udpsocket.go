package main

import (
    "net";
)

type UDPSocket struct {
    connection *net.UDPConn
}

func NewUDPSocket(address string) *UDPSocket {

    udpAddr, err := net.ResolveUDPAddr("udp4", address)
    if err != nil { panic(err) }

    udpConn, err := net.ListenUDP("udp4", udpAddr)
    if err != nil { panic(err) }

    return &UDPSocket{udpConn}
}

func (socket *UDPSocket) Receive() ([]byte, string, bool) {

    buffer := make([]byte, 1024)

    n, peer, err := socket.connection.ReadFromUDP(buffer)
    if err != nil { return []byte{}, "", false }

    return buffer[:n], peer.String(), true
}

func (socket *UDPSocket) Send(bytes []byte, address string) {

    udpAddr, err := net.ResolveUDPAddr("udp4", address)
    if err != nil { panic(err) }

    _, err = socket.connection.WriteToUDP(bytes, udpAddr)
    if err != nil { panic(err)  }

}

func (socket *UDPSocket) Unbind() {
    socket.connection.Close()
}