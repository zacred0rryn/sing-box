package sniff

import (
	"context"
	"encoding/binary"
	"io"
	"os"
	"time"

	"github.com/sagernet/sing-box/adapter"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/buf"
	"github.com/sagernet/sing/common/task"

	mDNS "github.com/miekg/dns"
)

func StreamDomainNameQuery(readCtx context.Context, reader io.Reader) (*adapter.InboundContext, error) {
	var length uint16
	err := binary.Read(reader, binary.BigEndian, &length)
	if err != nil {
		return nil, err
	}
	if length == 0 {
		return nil, os.ErrInvalid
	}
	_buffer := buf.StackNewSize(int(length))
	defer common.KeepAlive(_buffer)
	buffer := common.Dup(_buffer)
	defer buffer.Release()

	readCtx, cancel := context.WithTimeout(readCtx, time.Millisecond*100)
	var readTask task.Group
	readTask.Append0(func(ctx context.Context) error {
		return common.Error(buffer.ReadFullFrom(reader, buffer.FreeLen()))
	})
	err = readTask.Run(readCtx)
	cancel()
	if err != nil {
		return nil, err
	}
	return DomainNameQuery(readCtx, buffer.Bytes())
}

func DomainNameQuery(ctx context.Context, packet []byte) (*adapter.InboundContext, error) {
	var msg mDNS.Msg
	err := msg.Unpack(packet)
	if err != nil {
		return nil, err
	}
	return &adapter.InboundContext{Protocol: C.ProtocolDNS}, nil
}
