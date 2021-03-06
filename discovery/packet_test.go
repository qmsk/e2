package discovery

import (
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"net"
	"regexp"
	"testing"
)

func decodeTestPacket(str string) []byte {
	str = regexp.MustCompile(`\s+|--.+`).ReplaceAllString(str, "")

	if buf, err := hex.DecodeString(str); err != nil {
		panic(err)
	} else {
		return buf
	}
}

var testPacketBytes = decodeTestPacket(`
       68 6f 73 74 6e 61 6d 65 3d 45 32 3a 39 38 37 36
       3a 74 65 73 74 6f 72 67 5f 45 32 3a 30 3a 31 3a
       30 30 24 31 33 24 39 35 24 31 31 24 32 32 24 33
       33 3a 32 2e 38 2e 36 30 32 00 69 70 2d 61 64 64
       72 65 73 73 3d 31 39 32 2e 31 36 38 2e 30 2e 31
       37 36 00 6d 61 63 2d 61 64 64 72 65 73 73 3d 30
       30 3a 31 33 3a 39 35 3a 31 31 3a 32 32 3a 33 33
       00 74 79 70 65 3d 45 32 00
`)

var testPacket = Packet{
	Hostname:   "E2",
	XMLPort:    9876,
	Name:       "testorg_E2",
	UnitID:     0,
	VPCount:    1,
	MasterMac:  "00:13:95:11:22:33",
	Version:    "2.8.602",
	IPAddress:  "192.168.0.176",
	MacAddress: "00:13:95:11:22:33",
	Type:       "E2",
}

func testDecodePacket(t *testing.T, data []byte, expected Packet) {
	var udpAddr = &net.UDPAddr{IP: net.IP{127, 0, 0, 1}, Port: 1337}
	var actual Packet

	expected.IP = udpAddr.IP

	if err := actual.unpack(udpAddr, data); err != nil {
		assert.NoErrorf(t, err, "packet.unpack(...)")
	} else {
		assert.Equalf(t, expected, actual, "packet.unpack(...)")
	}
}

func testDecodePacketError(t *testing.T, data []byte, expected string) {
	var udpAddr = &net.UDPAddr{IP: net.IP{127, 0, 0, 1}, Port: 1337}
	var actual Packet

	err := actual.unpack(udpAddr, data)

	assert.EqualErrorf(t, err, expected, "packet.unpack(...)")
}

func TestDecodePacket(t *testing.T) {
	testDecodePacket(t, testPacketBytes, testPacket)
}

func TestDecodePacketInvalidSep(t *testing.T) {
	testDecodePacketError(t, []byte("foo\x00bar"), `Invalid field: []byte{0x66, 0x6f, 0x6f}`)
}
func TestDecodePacketInvalidHostname(t *testing.T) {
	testDecodePacketError(t, []byte("hostname=foo:bar"), `Invalid hostname="foo:bar": Invalid XMLPort="bar": expected integer`)
}

func TestDecodePacket_EC200(t *testing.T) {
	testDecodePacket(t, decodeTestPacket(`
		686f 7374
		6e61 6d65 3d45 432d 3230 303a 4e2f 413a
		5379 7374 656d 313a 303a 4e2f 413a 4e2f
		413a 352e 302e 3335 3437 3900 6970 2d61
		6464 7265 7373 3d31 3932 2e31 3638 2e30
		2e31 3830 006d 6163 2d61 6464 7265 7373
		3d30 303a 3062 3a61 623a 3938 3a62 613a
		6366 0074 7970 653d 4543 2d32 3030 00
	`), Packet{
		Hostname:   "EC-200",
		Name:       "System1",
		Version:    "5.0.35479",
		IPAddress:  "192.168.0.180",
		MacAddress: "00:0b:ab:98:ba:cf",
		Type:       "EC-200",
	})
}
