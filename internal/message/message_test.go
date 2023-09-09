package message

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSSLRequest(t *testing.T) {
	raw := make([]byte, 8)
	binary.BigEndian.PutUint32(raw, 8)
	binary.BigEndian.PutUint32(raw[4:], SSLRequestCode)

	reader := bytes.NewReader(raw)
	msg, err := ParseSSLRequest(reader)

	assert.Nil(t, err)
	assert.Equal(t, SSLRequest, msg.Type)
	assert.Equal(t, SSLRequestCode, binary.BigEndian.Uint32(msg.Data))
	assert.Equal(t, uint32(8), msg.Length)
}

func TestParseSSLRequestInvalidLength(t *testing.T) {
	raw := make([]byte, 6)
	binary.BigEndian.PutUint32(raw, 6)
	raw[4] = byte(0)
	raw[5] = byte(0)

	reader := bytes.NewReader(raw)
	msg, err := ParseSSLRequest(reader)

	assert.Nil(t, msg)
	assert.ErrorIs(t, err, ErrInvalidSSLRequestLength)
}

func TestParseSSLRequestInvalidRequestCode(t *testing.T) {
	raw := make([]byte, 8)
	binary.BigEndian.PutUint32(raw, 8)
	binary.BigEndian.PutUint32(raw[4:], uint32(1000))

	reader := bytes.NewReader(raw)
	msg, err := ParseSSLRequest(reader)

	assert.Nil(t, msg)
	assert.ErrorIs(t, err, ErrInvalidSSLRequestCode)
}

func TestParseSSLRequestBytes(t *testing.T) {
	expected := make([]byte, 8)
	binary.BigEndian.PutUint32(expected, 8)
	binary.BigEndian.PutUint32(expected[4:], SSLRequestCode)

	msg := NewSSLRequestMessage()

	assert.Equal(t, expected, msg.Bytes())
}

// TODO: Parse SSLResponse

func TestParseStartupMessage(t *testing.T) {
	raw := make([]byte, 8)
	binary.BigEndian.PutUint32(raw, 8)
	binary.BigEndian.PutUint32(raw[4:], ProtocolVersion)

	reader := bytes.NewReader(raw)
	msg, err := ParseStartupMessage(reader)

	assert.Nil(t, err)
	assert.Equal(t, StartupMessage, msg.Type)
	assert.Equal(t, ProtocolVersion, binary.BigEndian.Uint32(msg.Data))
	assert.Equal(t, uint32(8), msg.Length)
}

func TestParseStartupMessageInvalidLength(t *testing.T) {
	raw := make([]byte, 6)
	binary.BigEndian.PutUint32(raw, 6)
	raw[4] = 0
	raw[5] = 0

	reader := bytes.NewReader(raw)
	msg, err := ParseStartupMessage(reader)

	assert.Nil(t, msg)
	assert.ErrorIs(t, err, ErrInvalidStartupMessageLength)
}

func TestParseStartupMessageInvalidProtocolVersion(t *testing.T) {
	raw := make([]byte, 8)
	binary.BigEndian.PutUint32(raw, 8)
	binary.BigEndian.PutUint32(raw[4:], 0)

	reader := bytes.NewReader(raw)
	msg, err := ParseStartupMessage(reader)

	assert.Nil(t, msg)
	assert.ErrorIs(t, err, ErrUnsupportedProtocolVersion)
}

func TestStartupMessageBytes(t *testing.T) {
	expected := make([]byte, 8)
	binary.BigEndian.PutUint32(expected, 8)
	binary.BigEndian.PutUint32(expected[4:], ProtocolVersion)

	msg := NewStartupMessage(ProtocolVersion)

	assert.Equal(t, expected, msg.Bytes())
}

func TestParseMessage(t *testing.T) {
	var raw bytes.Buffer

	data := "Some other arbitrary data"
	dataBytes := []byte(data)
	dataLen := uint32(len(dataBytes))

	msgLen := dataLen + MessageLengthByteSize
	msgLenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(msgLenBytes, uint32(msgLen))

	raw.WriteString("R")
	raw.Write(msgLenBytes)
	raw.WriteString(data)

	reader := bytes.NewReader(raw.Bytes())
	msg, err := ParseMessage(reader)

	assert.Nil(t, err)
	assert.Equal(t, AuthenticationRequest, msg.Type)
	assert.Equal(t, msgLen, msg.Length)
	assert.Equal(t, data, string(msg.Data))
}

func TestParseMessageInvalidLength(t *testing.T) {
	var raw bytes.Buffer

	msgLenBytes := make([]byte, MessageLengthByteSize)
	binary.BigEndian.PutUint32(msgLenBytes, uint32(2))

	raw.WriteString("R")
	raw.Write(msgLenBytes)

	reader := bytes.NewReader(raw.Bytes())
	msg, err := ParseMessage(reader)

	assert.Nil(t, msg)
	assert.ErrorIs(t, err, ErrInvalidMessageLength)
}

func TestMessageBytes(t *testing.T) {
	len := 1234
	data := make([]byte, len)
	msg := NewMessage(AuthenticationRequest, data)
	bytes := msg.Bytes()

	assert.Equal(t, string(AuthenticationRequest), string(bytes[0]))
	assert.Equal(t, uint32(len)+MessageLengthByteSize, binary.BigEndian.Uint32(bytes[1:5]))
	assert.Equal(t, data, msg.Data)
}
