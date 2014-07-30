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
	timeout   = flag.Uint("timeout", 10, "")
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()

	restapi.NewServer(restapi.Options{
		HttpAddr:  *httpAddr,
		Increment: *increment,
		Offset:    *offset,
		DataDir:   *dataDir,
		LogDir:    *logDir,
		Timeout:   *timeout,
	})
}
