package restapi

import (
	"errors"
	"fmt"
	"github.com/kshvakov/sequence-generator/generator"
	"log"
	"strconv"
	"strings"
	"time"
	//"os"
	"io/ioutil"
	"net/http"
	//	_ "net/http/pprof" // debug
)

var (
	serverLog         *log.Logger
	sequenceGenerator *generator.SequenceGenerator
	stat              *Stat
)

func sequence(writer http.ResponseWriter, request *http.Request) {

	switch request.Method {

	case "GET":

		stat.add("get")

		key, err := getKey(request.URL.Path)

		if err != nil {

			log.Print(err.Error())

			sendErrorResponse(writer, err.Error(), http.StatusBadRequest)

			return
		}

		if value, err := sequenceGenerator.Get(key); err == nil {

			fmt.Fprint(writer, response{Key: key, Value: value})
		} else {

			log.Print(err.Error())

			sendErrorResponse(writer, "500 Internal Server Error", http.StatusInternalServerError)
		}

	case "PUT":

		stat.add("add")

		key, err := getKey(request.URL.Path)

		if err != nil {

			log.Print(err.Error())

			sendErrorResponse(writer, err.Error(), http.StatusBadRequest)

			return
		}

		body, err := ioutil.ReadAll(request.Body)

		if err != nil {

			log.Print(err.Error())

			sendErrorResponse(writer, err.Error(), http.StatusBadRequest)

			return
		}

		value, err := strconv.ParseUint(string(body), 10, 0)

		if err != nil {

			log.Print(err.Error())

			sendErrorResponse(writer, err.Error(), http.StatusBadRequest)

			return
		}

		if err := sequenceGenerator.Add(key, uint(value)); err != nil {
			log.Print(err.Error())

			sendErrorResponse(writer, err.Error(), http.StatusBadRequest)

			return
		}

		fmt.Fprint(writer, response{Key: key, Value: uint(value)})

	default:
		sendErrorResponse(writer, "405 method not allowed", http.StatusMethodNotAllowed)
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

func statistics(writer http.ResponseWriter, request *http.Request) {

	if result, err := stat.getStat(); err == nil {

		fmt.Fprint(writer, result)
	} else {

		log.Print(err.Error())

		sendErrorResponse(writer, "500 Internal Server Error", http.StatusInternalServerError)
	}
}

func NewServer(httpAddr string, increment uint, offset uint, dataDir string, logDir string) {

	stat = NewStat()

	sequenceGenerator = generator.NewGenerator(generator.Options{
		Increment: increment,
		Offset:    offset,
		DataDir:   dataDir,
		LogDir:    logDir,
	})

	httpServer := &http.Server{
		Addr:         httpAddr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		sendErrorResponse(writer, "404 method not found", http.StatusNotFound)
	})

	http.HandleFunc("/ping/", func(writer http.ResponseWriter, request *http.Request) {

		stat.add("ping")

		fmt.Fprint(writer, "pong")
	})

	http.HandleFunc("/stat/", handle(statistics))
	http.HandleFunc("/sequence/", handle(sequence))

	log.Fatal(httpServer.ListenAndServe())
}

func handle(function func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {

	return func(writer http.ResponseWriter, request *http.Request) {

		defer func() {

			if panic := recover(); panic != nil {

				log.Printf("panic: %v", panic)

				sendErrorResponse(writer, "500 Internal Server Error", http.StatusInternalServerError)
			}
		}()

		function(writer, request)
	}
}

func sendErrorResponse(writer http.ResponseWriter, errorString string, code int) {

	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(code)

	fmt.Fprint(writer, response{Error: errorString})
}
