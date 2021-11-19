package proxy_test

import (
	"bytes"
	"fmt"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

var TerminateTest = fmt.Errorf("terminate test")

type ControlPacketReader struct {
	ControlPacketList []packets.ControlPacket
	index             int
	buf               *bytes.Buffer
}

func (r *ControlPacketReader) Read(p []byte) (n int, err error) {
	if r.buf == nil {
		if len(r.ControlPacketList) <= r.index {
			return 0, TerminateTest
		}
		r.buf = &bytes.Buffer{}
		cp := r.ControlPacketList[r.index]
		r.index = r.index + 1

		if err = cp.Write(r.buf); err != nil {
			return 0, err
		}
	}

	n, err = r.buf.Read(p)

	if r.buf.Len() == 0 {
		r.buf = nil
	}

	return n, err
}
