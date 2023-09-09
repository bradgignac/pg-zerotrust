package message

import (
	"encoding/binary"
	"fmt"
	"io"
)

const MessageTypeFieldSize = uint32(1)
const MessageLengthByteSize = uint32(4)
const ProtocolVersion = uint32(196608)
const SSLRequestCode = uint32(80877103)

var ErrInvalidMessageLength = fmt.Errorf("Message must be at least 4-bytes long")
var ErrInvalidSSLRequestCode = fmt.Errorf("SSLRequest code must be %d", SSLRequestCode)
var ErrInvalidSSLRequestLength = fmt.Errorf("SSLRequest message must be 8-bytes long")
var ErrInvalidStartupMessageLength = fmt.Errorf("StartupMessage must be at least 8-bytes long")
var ErrUnsupportedProtocolVersion = fmt.Errorf("Only protocol version %d is supported", ProtocolVersion)

type MessageType string

const (
	StartupMessage        MessageType = "StartupMessage"
	SSLRequest            MessageType = "SSLRequest"
	AuthenticationRequest MessageType = "R"
)

type Message struct {
	Type   MessageType
	Length uint32
	Data   []byte
}

func NewMessage(msgType MessageType, data []byte) *Message {
	len := MessageLengthByteSize + uint32(len(data))
	msg := &Message{Type: msgType, Data: data, Length: len}

	return msg
}

func NewSSLRequestMessage() *Message {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, SSLRequestCode)

	return NewMessage(SSLRequest, data)
}

func NewStartupMessage(version uint32) *Message {
	// TODO: Support key-value params

	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, version)

	return NewMessage(StartupMessage, data)
}

func ParseSSLRequest(r io.Reader) (*Message, error) {
	const expectedMsgLen = 8

	msgLen, err := getMessageLength(r)
	if err != nil {
		return nil, err
	} else if msgLen != expectedMsgLen {
		return nil, ErrInvalidSSLRequestLength
	}

	msgData, err := getMessageData(r, msgLen-MessageLengthByteSize)
	if err != nil {
		return nil, err
	} else if binary.BigEndian.Uint32(msgData) != SSLRequestCode {
		return nil, ErrInvalidSSLRequestCode
	}

	msg := &Message{Type: SSLRequest, Length: msgLen, Data: msgData}
	return msg, nil
}

func ParseStartupMessage(r io.Reader) (*Message, error) {
	const minMsgLen = 8

	msgLen, err := getMessageLength(r)
	if err != nil {
		return nil, err
	} else if msgLen < minMsgLen {
		return nil, ErrInvalidStartupMessageLength
	}

	msgData, err := getMessageData(r, msgLen-MessageLengthByteSize)
	if err != nil {
		return nil, err
	} else if binary.BigEndian.Uint32(msgData) != ProtocolVersion {
		return nil, ErrUnsupportedProtocolVersion
	}

	msg := &Message{Type: StartupMessage, Length: msgLen, Data: msgData}
	return msg, nil
}

func ParseMessage(r io.Reader) (*Message, error) {
	const minMsgLen = 4

	msgType, err := getMessageType(r)
	if err != nil {
		return nil, err
	}

	msgLen, err := getMessageLength(r)
	if err != nil {
		return nil, err
	} else if msgLen < minMsgLen {
		return nil, ErrInvalidMessageLength
	}

	msgData, err := getMessageData(r, msgLen-MessageLengthByteSize)
	if err != nil {
		return nil, err
	}

	msg := &Message{Type: msgType, Length: msgLen, Data: msgData}
	return msg, nil
}

func (m *Message) Bytes() []byte {
	buffer := make([]byte, MessageLengthByteSize)
	binary.BigEndian.PutUint32(buffer, m.Length)
	buffer = append(buffer, m.Data...)

	switch m.Type {
	case SSLRequest:
		fallthrough
	case StartupMessage:
		return buffer
	}

	return append([]byte(m.Type), buffer...)
}

func getMessageType(r io.Reader) (MessageType, error) {
	buffer := make([]byte, MessageTypeFieldSize)
	_, err := r.Read(buffer)
	if err != nil {
		return "", err
	}

	msgType := string(buffer)
	return MessageType(msgType), nil
}

func getMessageLength(r io.Reader) (uint32, error) {
	buffer := make([]byte, MessageLengthByteSize)
	_, err := r.Read(buffer)
	if err != nil {
		return 0, err
	}

	msgLen := binary.BigEndian.Uint32(buffer)
	return msgLen, nil
}

func getMessageData(r io.Reader, len uint32) ([]byte, error) {
	buffer := make([]byte, len)
	_, err := r.Read(buffer)

	return buffer, err
}
