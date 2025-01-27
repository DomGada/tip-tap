package client

import (
	"github.com/DomGada/Tip-Tap/shared"
	"net"
)

type TipTapClient struct {
	connection net.Conn
	logger     shared.Logger
}
