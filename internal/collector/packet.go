package collector

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/andybalholm/brotli"
)

const (
	OpHeartbeatReply   = 3
	OpHeartbeat        = 2
	OpServerMessage    = 5
	OpAuth             = 7
	OpAuthReply        = 8
	ProtoJSON          = 0
	ProtoHeartbeat     = 1
	ProtoBrotli        = 3
	headerBytes        = 16
)

type Packet struct {
	TotalLen  uint32
	HeaderLen uint16
	ProtoVer  uint16
	Op        uint32
	Seq       uint32
	Header    []byte
	Body      []byte
}

func ParsePacket(r io.Reader) (*Packet, error) {
	buf := make([]byte, headerBytes)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	p := &Packet{
		TotalLen:  binary.BigEndian.Uint32(buf[0:4]),
		HeaderLen: binary.BigEndian.Uint16(buf[4:6]),
		ProtoVer:  binary.BigEndian.Uint16(buf[6:8]),
		Op:        binary.BigEndian.Uint32(buf[8:12]),
		Seq:       binary.BigEndian.Uint32(buf[12:16]),
	}
	bodyLen := int(p.TotalLen) - headerBytes
	rest := make([]byte, bodyLen)
	if _, err := io.ReadFull(r, rest); err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	if int(p.HeaderLen) > 0 && int(p.HeaderLen) <= len(rest) {
		p.Header = rest[:p.HeaderLen]
	}
	if bodyLen > int(p.HeaderLen) {
		p.Body = rest[p.HeaderLen:]
	}
	if p.ProtoVer == ProtoBrotli && len(p.Body) > 0 {
		decompressed, err := decompressBrotli(p.Body)
		if err != nil {
			return nil, fmt.Errorf("brotli: %w", err)
		}
		p.Body = decompressed
	}
	return p, nil
}

func decompressBrotli(data []byte) ([]byte, error) {
	r := brotli.NewReader(bytes.NewReader(data))
	var out bytes.Buffer
	if _, err := io.Copy(&out, r); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

type PacketWriter struct {
	w io.Writer
}

func NewPacketWriter(w io.Writer) *PacketWriter {
	return &PacketWriter{w: w}
}

func (pw *PacketWriter) Write(op uint32, body []byte) error {
	totalLen := headerBytes + len(body)
	buf := make([]byte, totalLen)
	binary.BigEndian.PutUint32(buf[0:4], uint32(totalLen))
	binary.BigEndian.PutUint16(buf[4:6], 0) // no additional header for JSON packets
	binary.BigEndian.PutUint16(buf[6:8], ProtoJSON)
	binary.BigEndian.PutUint32(buf[8:12], op)
	binary.BigEndian.PutUint32(buf[12:16], 1)
	copy(buf[16:], body)
	_, err := pw.w.Write(buf)
	return err
}
