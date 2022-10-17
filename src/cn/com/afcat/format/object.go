package format

import (
	"bytes"
	"encoding/binary"
	"github.com/go-netty/go-netty"
	"github.com/go-netty/go-netty/codec"
	"github.com/go-netty/go-netty/utils"
	"io/ioutil"
	"strings"
)

// TextCodec create a text codec
func ObjectCodec() codec.Codec {
	return &objectCodec{}
}

type objectCodec struct{}

func (*objectCodec) CodecName() string {
	return "object-codec"
}

func (*objectCodec) HandleRead(ctx netty.InboundContext, message netty.Message) {

	// read text bytes
	textBytes := utils.MustToBytes(message)

	// convert from []byte to string
	sb := strings.Builder{}
	sb.Write(textBytes)

	// post text
	ctx.HandleRead(sb.String())
}

func (*objectCodec) HandleWrite(ctx netty.OutboundContext, message netty.Message) {

	switch s := message.(type) {
	case string:
		buff := make([]byte, 4)
		buff[0] = 5
		buff[1] = 116
		binary.BigEndian.PutUint16(buff[2:4], uint16(len(s)))
		//binary.BigEndian.PutUint16(buff[:2], STREAM_MAGIC)
		//binary.BigEndian.PutUint16(buff[:1], STREAM_VERSION)
		readerMess := strings.NewReader(s)
		all, _ := ioutil.ReadAll(readerMess)
		sendMess := append(buff, all...)
		reader := bytes.NewReader(sendMess)
		ctx.HandleWrite(reader)
	default:
		ctx.HandleWrite(message)
	}
}
