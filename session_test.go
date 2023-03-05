/**
 * @Author: koulei
 * @Description:
 * @File: session_test
 * @Version: 1.0.0
 * @Date: 2023/3/5 00:07
 */

package link

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
)

type ProtocolTest struct {
}

type ProtocolCodec struct {
	w      io.Writer
	r      io.Reader
	closer io.Closer

	buffReceiving *bytes.Buffer
}

func (codec *ProtocolCodec) Send(msg interface{}) error {
	marshal, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, _ = codec.w.Write(marshal)
	return nil
}

func (codec *ProtocolCodec) Receive() (interface{}, error) {
	msg, ok, err := codec.readFromBuff()
	if ok {
		return msg, err
	}

	if err != nil {
		return nil, err
	}

	var buff [128]byte
	for {
		count, err := io.ReadAtLeast(codec.r, buff[:], 1)
		if err != nil {
			return nil, err
		}

		codec.buffReceiving.Write(buff[:count])
		if codec.buffReceiving.Len() == 0 {
			continue
		}
		fmt.Println(codec.buffReceiving.Len(), 0xffff)
		if codec.buffReceiving.Len() > 256 {
			return nil, errors.New("数据太长")
		}

		msg, ok, err = codec.readFromBuff()
		if ok {
			return msg, nil
		}
		if err != nil {
			return nil, err
		}
	}
}

type message struct {
	name string
}

func (codec *ProtocolCodec) readFromBuff() (interface{}, bool, error) {
	if codec.buffReceiving.Len() == 0 {
		return nil, false, nil
	}
	data := codec.buffReceiving.Bytes()
	if string(data[0]) != "s" {
		i := 0
		for ; i < len(data); i++ {
			if string(data[i]) == "s" {
				break
			}
		}
		codec.buffReceiving.Next(i)
		return nil, false, errors.New("无效的数据头")
	}

	end := 1
	for ; end < len(data); end++ {
		if string(data[end]) == "s" {
			break
		}
	}
	if end == len(data) {
		return nil, false, errors.New("无效的数据尾")
	}
	codec.buffReceiving.Next(end + 1)
	return string(data[1:end]), true, nil
}

func (codec *ProtocolCodec) Close() error {
	if codec.closer != nil {
		return codec.closer.Close()
	}
	return nil
}

type sessionHandler struct {
}

func (handler sessionHandler) HandlerSession(session *Session) {
	for {
		msg, err := session.codec.Receive()
		if err != nil {
			session.Close()
			fmt.Println(err.Error())
			return
		}
		if err = session.Send(msg); err != nil {
			panic(err)
		}
		continue
	}
}

func (p *ProtocolTest) NewCodec(rw io.ReadWriter) (Codec, error) {
	codec := &ProtocolCodec{
		w:             rw,
		r:             rw,
		buffReceiving: bytes.NewBuffer(nil),
	}
	closer, ok := rw.(io.Closer)
	if ok {
		codec.closer = closer
	}
	return codec, nil
}

func TestRemainder(t *testing.T) {
	listener, err := net.Listen("tcp", "0.0.0.0:8109")
	if err != nil {
		panic(err)
	}

	p := &ProtocolTest{}
	server := NewServer(listener, p, 0, sessionHandler{})
	if err = server.Serve(); err != nil {
		panic(err)
	}
}
