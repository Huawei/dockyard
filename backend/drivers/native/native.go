package native

import (
	"github.com/containerops/dockyard/backend/drivers"
)

func init() {
	drivers.Register("native", InitFunc)
}

func InitFunc() {

}
