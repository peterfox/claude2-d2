package r2

import "sync/atomic"

const (
	sopByte = 0x8D
	eopByte = 0xD8
	escByte = 0xAB

	escEsc = 0x23
	escSop = 0x05
	escEop = 0x50
)

var globalSeq atomic.Uint32

func nextSeq() byte {
	return byte(globalSeq.Add(1) % 140)
}

func buildPacket(init []byte, payload []byte) []byte {
	seq := nextSeq()

	body := make([]byte, 0, len(init)+1+len(payload)+1)
	body = append(body, init...)
	body = append(body, seq)
	body = append(body, payload...)

	sum := 0
	for _, b := range body {
		sum += int(b)
	}
	checksum := byte(0xFF - (sum & 0xFF))

	body = append(body, checksum)
	return encode(body)
}

func encode(raw []byte) []byte {
	packet := []byte{sopByte}
	for _, b := range raw {
		switch b {
		case escByte:
			packet = append(packet, escByte, escEsc)
		case sopByte:
			packet = append(packet, escByte, escSop)
		case eopByte:
			packet = append(packet, escByte, escEop)
		default:
			packet = append(packet, b)
		}
	}
	packet = append(packet, eopByte)
	return packet
}
