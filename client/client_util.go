package client

import "github.com/DomGada/Tip-Tap/shared"

func makeEmptyClient() TipTapClient {
	return TipTapClient{nil, shared.Logger{}}
}
