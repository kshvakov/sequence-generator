package generator

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"github.com/kshvakov/hlp/fs"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

var (
	errorLog  *log.Logger
	infoLog   *log.Logger
	systemLog *log.Logger
)

// SequenceGenerator
type SequenceGenerator struct {
	counter     sequenceCounter
	actionMutex *sync.RWMutex

	incrementIncrement uint
	incrementOffset    uint

	dataLogFile    *os.File
	dataLogDirPath string

	dataFilePath     string
	snapshotFilePath string

	saveDataMutex *sync.RWMutex

	flushChan      chan string
	logSizeCounter int
	currentLogNum  int
}

func (sequence *SequenceGenerator) new() {

	infoLog.Printf("Start. Pid: %d", os.Getpid())

	signalChan := make(chan os.Signal)

	signal.Notify(signalChan,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGKILL,
	)

	go func() {

		<-signalChan

		if err := sequence.stop(); err != nil {

			errorLog.Print(err.Error())
		}

		os.Exit(1)
	}()

	go sequence.flushLog()

	sequence.load()
}

func (sequence *SequenceGenerator) load() {

	if fs.IsNotExist(sequence.dataLogDirPath) {

		infoLog.Printf("%s: %s", "Create log dir", sequence.dataLogDirPath)

		err := os.Mkdir(sequence.dataLogDirPath, permissions)

		if err != nil {

			errorLog.Fatal(err.Error())
		}
	}

	var (
		err      error
		dataFile *os.File
	)

	if fs.IsExist(sequence.dataFilePath) {

		dataFile, err = os.Open(sequence.dataFilePath)

	} else {

		infoLog.Printf("Create data file: %s", sequence.dataFilePath)

		dataFile, err = os.Create(sequence.dataFilePath)
	}

	defer dataFile.Close()

	if err != nil {

		errorLog.Fatal(err.Error())
	}

	decoder := gob.NewDecoder(dataFile)

	if err := decoder.Decode(&sequence.counter); err != nil {

		errorLog.Fatal(err.Error())
	}

	logFiles, err := ioutil.ReadDir(sequence.dataLogDirPath)

	if err != nil {

		errorLog.Fatal(err.Error())
	}

	for _, logFileInfo := range logFiles {

		logFilePath := sequence.dataLogDirPath + logFileInfo.Name()

		if !logFileInfo.IsDir() && strings.HasSuffix(logFilePath, ".log") {

			infoLog.Printf("Read log: %s", logFilePath)

			logFile, err := os.Open(logFilePath)

			if err != nil {

				errorLog.Fatal(err.Error())
			}

			defer logFile.Close()

			scanner := bufio.NewScanner(logFile)

			var key string
			var value uint

			for scanner.Scan() {

				if _, err := fmt.Sscanf(scanner.Text(), dataLogFormat, &key, &value); err == nil {

					if val, exist := sequence.counter[key]; !exist || val < value {
						sequence.counter[key] = value
					}
				}
			}

			os.Remove(logFilePath)
		}
	}

	filename := sequence.getCurrentLogPath()

	infoLog.Printf("write log: %s", filename)

	if file, err := os.Create(filename); err == nil {

		sequence.dataLogFile = file

		sequence.saveDataFile()
	} else {

		errorLog.Fatal(err.Error())
	}
}

// Get next increment of the key
func (sequence *SequenceGenerator) Get(key string) (uint, error) {

	sequence.actionMutex.Lock()

	defer sequence.actionMutex.Unlock()

	sequence.counter[key]++

	if err := sequence.addLog(key, sequence.counter[key]); err != nil {

		errorLog.Print(err.Error())

		return 0, err
	}

	return offset(sequence.counter[key], sequence.incrementIncrement, sequence.incrementOffset), nil
}

// Add new key (with value) to storage
func (sequence *SequenceGenerator) Add(key string, value uint) error {

	sequence.actionMutex.Lock()

	defer sequence.actionMutex.Unlock()

	if _, exists := sequence.counter[key]; exists {

		return fmt.Errorf("key \"%s\" is exists", key)
	}

	sequence.counter[key] = value

	if err := sequence.addLog(key, value); err != nil {

		errorLog.Print(err.Error())

		return err
	}

	return nil
}

// Len of all keys
func (sequence *SequenceGenerator) Len() int {

	sequence.actionMutex.Lock()

	defer sequence.actionMutex.Unlock()

	return len(sequence.counter)
}

func (sequence *SequenceGenerator) addLog(key string, value uint) error {

	if sequence.logSizeCounter >= maxDataLogSize {

		if err := sequence.rotateLog(); err != nil {

			return err
		}
	}

	if _, err := io.WriteString(sequence.dataLogFile, fmt.Sprintf(dataLogFormat+"\n", key, value)); err != nil {

		errorLog.Print(err.Error())

		return fmt.Errorf("%s: %s", "Error write to log", err.Error())
	}

	sequence.logSizeCounter++

	return nil
}

func (sequence *SequenceGenerator) rotateLog() error {

	sequence.currentLogNum++
	sequence.logSizeCounter = 0

	currentLog := sequence.dataLogFile.Name()

	if err := sequence.dataLogFile.Close(); err != nil {

		errorLog.Print(err.Error())

		return err
	}

	filename := sequence.getCurrentLogPath()

	systemLog.Printf("write log: %s", filename)

	file, err := os.Create(filename)

	if err != nil {

		errorLog.Print(err.Error())

		return err
	}

	sequence.flushChan <- currentLog

	sequence.dataLogFile = file

	return nil
}

func (sequence *SequenceGenerator) flushLog() {

	for logFilePath := range sequence.flushChan {

		systemLog.Printf("Rotate: %s", logFilePath)

		if err := sequence.saveDataFile(); err == nil {

			os.Remove(logFilePath)
		}
	}
}

func (sequence *SequenceGenerator) saveDataFile() error {

	sequence.saveDataMutex.Lock()

	defer sequence.saveDataMutex.Unlock()

	tmpDataFilePath := sequence.dataFilePath + ".tmp"

	file, err := os.Create(tmpDataFilePath)

	defer file.Close()

	if err != nil {

		errorLog.Print(err.Error())

		return err
	}

	encoder := gob.NewEncoder(file)

	if err := encoder.Encode(sequence.counter); err != nil {

		errorLog.Print(err.Error())

		return err
	}

	if err := os.Rename(tmpDataFilePath, sequence.dataFilePath); err != nil {

		errorLog.Print(err.Error())

		return err
	}

	return nil
}

func (sequence *SequenceGenerator) stop() error {

	infoLog.Print("Stop storage. Create snapshot file")

	file, err := os.Create(sequence.snapshotFilePath)

	defer file.Close()

	if err != nil {

		errorLog.Print(err.Error())

		return err
	}

	encoder := gob.NewEncoder(file)

	if err := encoder.Encode(sequence.counter); err != nil {

		errorLog.Print(err.Error())

		return err
	}

	return sequence.dataLogFile.Close()
}

func (sequence *SequenceGenerator) getCurrentLogPath() string {
	return fmt.Sprintf("%slog_%d.log", sequence.dataLogDirPath, sequence.currentLogNum)
}

func offset(value uint, increment uint, offset uint) uint {
	return offset + (value * increment)
}

// NewGenerator constructor
//
//  generator := NewGenerator(Options{
//	Increment: 1,
//	Offset:    0,
//	DataDir:   "/var/sequence-generator/data/",
//	LogDir:    "/var/log/sequence-generator/",
//  })
//
func NewGenerator(options Options) *SequenceGenerator {

	file, _ := os.OpenFile(options.LogDir+logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, permissions)

	mw := io.MultiWriter(file, os.Stdout)

	errorLog = log.New(mw, "Error: ", log.LstdFlags)

	infoLog = log.New(mw, "Info: ", log.LstdFlags)

	systemLog = log.New(file, "", log.LstdFlags)

	if options.Offset > options.Increment {

		errorLog.Fatalf("Offset can not be greater than Increment (%d > %d)", options.Offset, options.Increment)
	}

	if !fs.IsDir(options.DataDir) {

		errorLog.Fatalf("Data dir %s is not exist", options.DataDir)
	}

	if !fs.IsDir(options.LogDir) {

		errorLog.Fatalf("Log dir %s is not exist", options.LogDir)
	}

	if options.Increment == 0 {

		options.Increment = 1
		options.Offset = 0

		infoLog.Print("Set Increment=1, Offset=0")
	}

	generator := &SequenceGenerator{
		counter: make(sequenceCounter),

		incrementIncrement: options.Increment,
		incrementOffset:    options.Offset,

		dataFilePath:     options.DataDir + dataFileName,
		snapshotFilePath: options.DataDir + snapshotFileName,
		dataLogDirPath:   options.DataDir + "logs/",
		currentLogNum:    1,

		flushChan: make(chan string, 10),

		actionMutex:   new(sync.RWMutex),
		saveDataMutex: new(sync.RWMutex),
	}

	generator.new()

	return generator
}
