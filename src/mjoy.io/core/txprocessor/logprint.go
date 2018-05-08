package txprocessor


import (
	"fmt"
	"mjoy.io/log"
	"os"
)

var (
	logTag = "core.txprocessor"
	logger log.Logger
)

func init() {
	logger = log.GetLogger(logTag)
	if logger == nil {
		fmt.Errorf("Can not get logger(%s)\n", logTag)
		os.Exit(1)
	}
}
