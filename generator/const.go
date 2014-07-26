package generator

import (
	"os"
)

const (
	dataLogFormat    string      = "%s %d"
	maxDataLogSize   int         = 10000
	permissions      os.FileMode = 0755
	dataFileName                 = "data.gob"
	snapshotFileName             = "snapshot.gob"
	logFileName                  = "sequence-generator.log"
)
