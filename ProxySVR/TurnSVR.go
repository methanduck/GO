package ProxySVR

import (
	"github.com/pions/stun"
	"github.com/pions/turn"
)

func Start() {
	svr := new(turn.StartArguments)
	svr.Realm = "test"
	svr.UDPPort = 6866

	turn.Start(*svr)
}
