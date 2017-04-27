// divert.go - interface for OpenBSD divert(4) sockets.
//
// To the extent possible under law, Ivan Markin waived all copyright
// and related or neighboring rights to divert, using the creative
// commons "cc0" public domain dedication. See LICENSE or
// <http://creativecommons.org/publicdomain/zero/1.0/> for full details.

// +build openbsd

package divert

import (
	"errors"
	"net"
	"os"
	"strconv"
	"syscall"
)

const IPPROTO_DIVERT = 258

type Addr struct {
	Port uint16
}

func (a *Addr) Network() string {
	return "divert"
}

func (a *Addr) String() string {
	return strconv.FormatUint(uint64(a.Port), 16)
}

type Conn struct {
	addr *Addr
	fd   int
	sa   syscall.Sockaddr
}

func (c *Conn) Read(b []byte) (int, error) {
	n, _, err := syscall.Recvfrom(c.fd, b, 0)
	return n, err
}

func (c *Conn) Write(b []byte) (int, error) {
	err := syscall.Sendto(c.fd, b, 0, c.sa)
	if err != nil {
		return 0, os.NewSyscallError("socket", err)
	}
	return len(b), nil
}

func (c *Conn) Close() error {
	return syscall.Close(c.fd)
}

func Listen(network, address string) (*Conn, error) {
	if network != "divert" {
		return nil, errors.New("network is not divert")
	}
	s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, IPPROTO_DIVERT)
	if err != nil {
		return nil, os.NewSyscallError("socket", err)
	}
	port, err := strconv.ParseUint(address, 10, 16)
	if err != nil {
		return nil, net.InvalidAddrError("invalid divert port")
	}
	sa := &syscall.SockaddrInet4{Port: int(port)}
	if err := syscall.Bind(s, sa); err != nil {
		syscall.Close(s)
		return nil, os.NewSyscallError("bind", err)
	}
	return &Conn{fd: s, addr: &Addr{Port: uint16(port)}, sa: sa}, nil
}
