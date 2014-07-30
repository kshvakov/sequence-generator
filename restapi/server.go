package restapi

import (
	"errors"
	"fmt"
	"github.com/kshvakov/sequence-generator/generator"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	//	_ "net/http/pprof" // debug
)

const (
	defaultTimeout             = 10
	logFileName                = "rest-api-server.log"
	permissions    os.FileMode = 0750
)

var (
	sequenceGenerator *generator.SequenceGenerator
	systemLog         *log.Logger
	stat              *Stat
)

func NewServer(options Options) {

	if options.Timeout == 0 {
		options.Timeout = defaultTimeout
	}

	file, err := os.OpenFile(options.LogDir+logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, permissions)

	if err != nil {

		log.Fatalf("Culd not open log %s", options.LogDir+logFileName)
	} else {

		log.SetOutput(io.MultiWriter(os.Stdout, file))

		systemLog = log.New(file, "", log.LstdFlags)
	}

	stat = NewStat()

	sequenceGenerator = generator.NewGenerator(generator.Options{
		Increment: options.Increment,
		Offset:    options.Offset,
		DataDir:   options.DataDir,
		LogDir:    options.LogDir,
	})

	httpServer := &http.Server{
		Addr:         options.HttpAddr,
		ReadTimeout:  time.Duration(options.Timeout) * time.Second,
		WriteTimeout: time.Duration(options.Timeout) * time.Second,
	}

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		sendErrorResponse(writer, "404 method not found", http.StatusNotFound)
	})

	http.HandleFunc("/ping/", handle(func(writer http.ResponseWriter, request *http.Request) {

		stat.add("ping")

		fmt.Fprint(writer, "pong")
	}))

	http.HandleFunc("/stat/", handle(statistics))
	http.HandleFunc("/sequence/", handle(sequence))

	log.Printf("ListenAndServe on: %s", options.HttpAddr)

	log.Fatal(httpServer.ListenAndServe())
}

func sequence(writer http.ResponseWriter, request *http.Request) {

	key, err := getKey(request.URL.Path)

	if err != nil {

		sendErrorResponse(writer, err.Error(), http.StatusBadRequest)

		return
	}

	switch request.Method {

	case "GET":

		stat.add("get")

		if value, err := sequenceGenerator.Get(key); err == nil {

			fmt.Fprint(writer, response{Key: key, Value: value})
		} else {

			systemLog.Print(err.Error())

			sendInternalServerErrorResponse(writer)
		}

	case "PUT":

		stat.add("add")

		body, err := ioutil.ReadAll(request.Body)

		if err != nil {

			sendErrorResponse(writer, err.Error(), http.StatusBadRequest)

			return
		}

		value, err := strconv.ParseUint(string(body), 10, 0)

		if err != nil {

			sendErrorResponse(writer, err.Error(), http.StatusBadRequest)

			return
		}

		if err := sequenceGenerator.Add(key, uint(value)); err != nil {

			sendErrorResponse(writer, err.Error(), http.StatusBadRequest)

			return
		}

		writer.WriteHeader(http.StatusCreated)

		fmt.Fprint(writer, response{Key: key, Value: uint(value)})

	default:
		sendErrorResponse(writer, "405 method not allowed", http.StatusMethodNotAllowed)
	}
}

func statistics(writer http.ResponseWriter, request *http.Request) {

	if result, err := stat.getStat(); err == nil {

		fmt.Fprint(writer, result)
	} else {

		systemLog.Print(err.Error())

		sendInternalServerErrorResponse(writer)
	}
}

func handle(function func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {

	return func(writer http.ResponseWriter, request *http.Request) {

		defer func() {

			if panic := recover(); panic != nil {

				systemLog.Printf("panic: %v", panic)

				sendInternalServerErrorResponse(writer)
			}
		}()

		function(writer, request)
	}
}

func getKey(urlPath string) (string, error) {

	param := strings.Split(strings.Trim(urlPath, "/"), "/")

	if len(param) < 2 {

		return "", errors.New("missing key")

	} else if len(param) > 2 {

		return "", errors.New("too many parameters")
	}

	return param[1], nil
}

func sendErrorResponse(writer http.ResponseWriter, errorString string, code int) {

	systemLog.Print(errorString)

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(code)

	fmt.Fprint(writer, response{Error: errorString})
}

func sendInternalServerErrorResponse(writer http.ResponseWriter) {
	sendErrorResponse(writer, "500 Internal Server Error", http.StatusInternalServerError)
}
