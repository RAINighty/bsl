package collector

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/andybalholm/brotli"
)

// helper to build a B站 binary packet
func buildPacket(op uint32, protoVer uint16, header, body []byte) []byte {
	headerLen := uint16(len(header))
	totalLen := uint32(headerBytes + int(headerLen) + len(body))
	buf := make([]byte, totalLen)
	binary.BigEndian.PutUint32(buf[0:4], totalLen)
	binary.BigEndian.PutUint16(buf[4:6], headerLen)
	binary.BigEndian.PutUint16(buf[6:8], protoVer)
	binary.BigEndian.PutUint32(buf[8:12], op)
	binary.BigEndian.PutUint32(buf[12:16], 1)
	copy(buf[16:16+headerLen], header)
	copy(buf[16+headerLen:], body)
	return buf
}

// helper to brotli-compress data
func compressBrotli(data []byte) []byte {
	var buf bytes.Buffer
	w := brotli.NewWriter(&buf)
	_, _ = w.Write(data)
	_ = w.Close()
	return buf.Bytes()
}

func TestParsePacket_SimpleJSON(t *testing.T) {
	body := []byte(`{"cmd":"DANMU_MSG","info":["test"]}`)
	raw := buildPacket(OpServerMessage, ProtoJSON, nil, body)

	p, err := ParsePacket(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Op != OpServerMessage {
		t.Errorf("Op = %d, want %d", p.Op, OpServerMessage)
	}
	if p.ProtoVer != ProtoJSON {
		t.Errorf("ProtoVer = %d, want %d", p.ProtoVer, ProtoJSON)
	}
	if !bytes.Equal(p.Body, body) {
		t.Errorf("Body = %s, want %s", p.Body, body)
	}
	if len(p.Header) != 0 {
		t.Errorf("Header should be empty, got %d bytes", len(p.Header))
	}
}

func TestParsePacket_WithHeader(t *testing.T) {
	header := []byte(`{"ct":"test"}`)
	body := []byte(`{"cmd":"DANMU_MSG"}`)
	raw := buildPacket(OpServerMessage, ProtoJSON, header, body)

	p, err := ParsePacket(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(p.Header, header) {
		t.Errorf("Header = %s, want %s", p.Header, header)
	}
	if !bytes.Equal(p.Body, body) {
		t.Errorf("Body = %s, want %s", p.Body, body)
	}
}

func TestParsePacket_Brotli(t *testing.T) {
	body := []byte(`{"cmd":"DANMU_MSG","info":["弹幕内容"]}`)
	compressed := compressBrotli(body)
	raw := buildPacket(OpServerMessage, ProtoBrotli, nil, compressed)

	p, err := ParsePacket(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ProtoVer != ProtoBrotli {
		t.Errorf("ProtoVer = %d, want %d", p.ProtoVer, ProtoBrotli)
	}
	if !bytes.Equal(p.Body, body) {
		t.Errorf("Body = %s, want %s", p.Body, body)
	}
}

func TestParsePacket_HeartbeatReply(t *testing.T) {
	// Heartbeat reply has op=3, proto=heartbeat, body is popularity count (4 bytes BE)
	popularity := uint32(12345)
	body := make([]byte, 4)
	binary.BigEndian.PutUint32(body, popularity)
	raw := buildPacket(OpHeartbeatReply, ProtoHeartbeat, nil, body)

	p, err := ParsePacket(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Op != OpHeartbeatReply {
		t.Errorf("Op = %d, want %d", p.Op, OpHeartbeatReply)
	}
	if binary.BigEndian.Uint32(p.Body) != popularity {
		t.Errorf("popularity = %d, want %d", binary.BigEndian.Uint32(p.Body), popularity)
	}
}

func TestParsePacket_EmptyBody(t *testing.T) {
	raw := buildPacket(OpAuthReply, ProtoJSON, nil, nil)

	p, err := ParsePacket(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Op != OpAuthReply {
		t.Errorf("Op = %d, want %d", p.Op, OpAuthReply)
	}
	if len(p.Body) != 0 {
		t.Errorf("Body should be empty, got %d bytes", len(p.Body))
	}
}

func TestParsePacket_TruncatedHeader(t *testing.T) {
	raw := make([]byte, 10) // less than 16 header bytes
	_, err := ParsePacket(bytes.NewReader(raw))
	if err == nil {
		t.Fatal("expected error for truncated header, got nil")
	}
}

func TestParsePacket_TruncatedBody(t *testing.T) {
	// Declare totalLen larger than actual data
	buf := make([]byte, 20)
	binary.BigEndian.PutUint32(buf[0:4], 100) // claim 100 bytes total
	binary.BigEndian.PutUint16(buf[4:6], 0)   // header len 0
	binary.BigEndian.PutUint16(buf[6:8], ProtoJSON)
	binary.BigEndian.PutUint32(buf[8:12], OpServerMessage)
	binary.BigEndian.PutUint32(buf[12:16], 1)

	_, err := ParsePacket(bytes.NewReader(buf))
	if err == nil {
		t.Fatal("expected error for truncated body, got nil")
	}
}

func TestPacketWriter_Write(t *testing.T) {
	var buf bytes.Buffer
	pw := NewPacketWriter(&buf)
	body := []byte(`{"cmd":"AUTH"}`)
	if err := pw.Write(OpAuth, body); err != nil {
		t.Fatalf("write error: %v", err)
	}

	// Verify the written packet
	p, err := ParsePacket(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("parse written packet: %v", err)
	}
	if p.Op != OpAuth {
		t.Errorf("Op = %d, want %d", p.Op, OpAuth)
	}
	if !bytes.Equal(p.Body, body) {
		t.Errorf("Body = %s, want %s", p.Body, body)
	}
}
