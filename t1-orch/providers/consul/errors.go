package consul

import (
	"fmt"

	"github.com/pkg/errors"
)

//OpConsul consul operation
type OpConsul int

const (
	//OpKeys get keys
	OpKeys OpConsul = 1

	//OpGet get payload by key
	OpGet OpConsul = 2
)

//ErrConsul error from consul
type ErrConsul struct {
	error
	Path string
	Op   OpConsul
}

//errors
var (

	//ErrNotFound entity is not fount by Consul
	ErrNotFound = errors.New("not found")

	//ErrDataNotFit2Model when payload data from Consul are not fit to model struct
	ErrDataNotFit2Model = errors.New("data are not fot to model")
)

var op2s = map[OpConsul]string{
	OpKeys: "OpKeys",
	OpGet:  "OpGet",
}

//String impl fmt.Stringer
func (op OpConsul) String() string {
	return op2s[op]
}

//Error impl error i-face
func (e ErrConsul) Error() string {
	return fmt.Sprintf("Op:%s, Path:'%s', Err:%v", e.Op, e.Path, e.error)
}

//Cause errors.Causer
func (e ErrConsul) Cause() error {
	return e.error
}
