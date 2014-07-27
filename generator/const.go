package generator

import (
	"os"
)

const (
	dataLogFormat    string      = "%s %d"
	maxDataLogSize   int         = 10000
	permissions      os.FileMode = 0750
	dataFileName                 = "data.gob"
	snapshotFileName             = "snapshot.gob"
	logFileName                  = "sequence-generator.log"
)
