package providers

import (
	"context"
)

//GetOption used for Get
type GetOption interface {
	isGetOptionPrivate() //would never be implemented
}

//ServicesConfig abstract data type
type ServicesConfig interface {
	isBalancerDataPrivate() //would never be implemented
}

type ServicesConfigProvider interface {
	Get(ctx context.Context, opts ...GetOption) (ServicesConfig, error)
	Close() error
}
