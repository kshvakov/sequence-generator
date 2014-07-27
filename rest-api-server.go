package main

import (
	"flag"
	"github.com/kshvakov/sequence-generator/restapi"
	"runtime"
)

var (
	httpAddr  = flag.String("http", ":8080", "http=ip:port")
	logDir    = flag.String("log_dir", "/var/log/sequence-generator/", "")
	dataDir   = flag.String("data_dir", "/var/sequence-generator/", "usage")
	increment = flag.Uint("increment", 1, "")
	offset    = flag.Uint("offset", 0, "")
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()

	restapi.NewServer(*httpAddr, *increment, *offset, *dataDir, *logDir)
}
