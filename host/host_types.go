package host

import (
	"github.com/DomGada/Tip-Tap/shared"
	"net"
)

type TipTapHost struct {
	listener net.Listener
	logger   shared.Logger
}
