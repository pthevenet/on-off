// Package wol offers a function for sending a layer 2 (non ip) wake on lan packet
package wol

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"

	"github.com/mdlayher/ethernet"
	"github.com/mdlayher/raw"
)

///// API

// WakeOnLan sends a WoL packet (magic packet) to wake target.
// target is a MAC address that should be on the same ethernet LAN .
// iface is the interface to send the wol packet on.
func WakeOnLan(target net.HardwareAddr, iface *net.Interface) error {
	return NewMagicPacket(target).send(iface)
}

// NewMagicPacket constructs a MagicPacket with the target MAC address
func NewMagicPacket(target net.HardwareAddr) MagicPacket {
	var buf []byte
	head, err := hex.DecodeString("ffffffffffff")
	if err != nil {
		log.Panic("cannot decode string into bytes:", err)
	}
	buf = append(buf, head...)

	for i := 0; i < 16; i++ {
		buf = append(buf, []byte(target)...)
	}
	return MagicPacket{buf}
}

// MagicPacket is a broadcast frame containing FF FF FF FF FF FF followed by 16 repetitions of the target MAC address
type MagicPacket struct {
	payload []byte
}

///// Implementation

// Frame creates the ethernet frame for the magic packet
func (mpkt MagicPacket) frame() ethernet.Frame {
	frame := ethernet.Frame{
		Destination: ethernet.Broadcast,
		Source:      net.HardwareAddr{0xde, 0xad, 0xbe, 0xef, 0xde, 0xad},
		EtherType:   0x0842,
		Payload:     mpkt.payload,
	}
	return frame
}

// Send broadcasts a MagicPacket on given interface
func (mpkt MagicPacket) send(iface *net.Interface) error {
	conn, err := raw.ListenPacket(iface, 0x0842, nil)
	if err != nil {
		return fmt.Errorf("could not open ethernet socket: %v", err)
	}
	defer conn.Close()

	frame := mpkt.frame()

	b, err := frame.MarshalBinary()
	if err != nil {
		return fmt.Errorf("could not marshal frame: %v", err)
	}

	addr := &raw.Addr{HardwareAddr: ethernet.Broadcast}
	_, err = conn.WriteTo(b, addr)

	if err != nil {
		return fmt.Errorf("could not write to ethernet socket: %v", err)
	}
	return nil
}
