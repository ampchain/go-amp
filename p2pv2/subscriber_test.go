package p2pv2

import (
	"testing"

	"github.com/ampchain/go-amp/p2pv2/pb"
)

func TestNewSubscriber(t *testing.T) {
	ms := newMultiSubscriber()
	resch := make(chan *amperp2p.AmperMessage, 1)
	sub := NewSubscriber(resch, amperp2p.AmperMessage_PING, nil, "")
	sub, _ = ms.register(sub)
	if ms.elem.Len() != 1 {
		t.Error("register sub error")
	}
	ms.unRegister(sub)
	if ms.elem.Len() != 0 {
		t.Error("unRegister sub error")
	}
}
