/*
 * @Author: your name
 * @Date: 2021-03-25 13:41:11
 * @LastEditTime: 2021-03-26 15:51:26
 * @LastEditors: Please set LastEditors
 * @Description: In User Settings Edit
 * @FilePath: /socker/utils.go
 */
package socker

import (
	"bytes"
	"encoding/binary"

	"github.com/gofrs/uuid"
)

// 大端序字节转 int
func BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)

	var x int32
	_ = binary.Read(bytesBuffer, binary.BigEndian, &x)

	return int(x)
}

// int 转大端序字节
func IntToBytes(n int) []byte {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	_ = binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

func MergeBytes(b1 []byte, b2 ...[]byte) []byte {
	var buffer bytes.Buffer
	buffer.Write(b1)
	for _, b := range b2 {
		buffer.Write(b)
	}
	return buffer.Bytes()
}

func ProcessMessages(frame []byte) []Message {
	messages := make([]Message, 0)
	l := BytesToInt(frame[:4])
	for len(frame) > 4 {
		var message Message
		err := message.Parse(frame[4 : 4+l])
		if err != nil {
			message.err = err
		}
		messages = append(messages, message)
		frame = frame[4+l:]
	}
	return messages
}

func GenUUIDStr() string {
	UUID, _ := uuid.NewV4()
	return UUID.String()
}
