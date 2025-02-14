package magicbytes

import (
	"fmt"

	"github.com/Indra-Labs/indra"
	log2 "github.com/cybriq/proc/pkg/log"
)

var (
	log   = log2.GetLogger(indra.PathBase)
	check = log.E.Chk
)

const (
	Len         = 2
	ErrTooShort = "'%s' message  minimum size: %d got: %d"
)

func TooShort(got, found int, magic string) (e error) {
	e = fmt.Errorf(ErrTooShort, magic, got, found)
	check(e)
	return

}
