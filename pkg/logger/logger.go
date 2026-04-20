package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var (
	logLevelNames = map[LogLevel]string{
		DEBUG: "DEBUG",
		INFO:  "INFO",
		WARN:  "WARN",
		ERROR: "ERROR",
		FATAL: "FATAL",
	}

	currentLevel = INFO
	logger       *Logger
	once         sync.Once
	mu           sync.RWMutex
)

type Logger struct {
	file     *os.File
	logDir   string
	lastDate string
	prefix   string
}

type LogEntry struct {
	Level     string                 `json:"level"`
	Timestamp string                 `json:"timestamp"`
	Component string                 `json:"component,omitempty"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Caller    string                 `json:"caller,omitempty"`
}

func init() {
	once.Do(func() {
		logger = &Logger{}
	})
}

func SetLevel(level LogLevel) {
	mu.Lock()
	defer mu.Unlock()
	currentLevel = level
}

func GetLevel() LogLevel {
	mu.RLock()
	defer mu.RUnlock()
	return currentLevel
}

func EnableFileLogging(filePath string) error {
	mu.Lock()
	defer mu.Unlock()

	return enableFileLoggingLocked(filePath)
}

func enableFileLoggingLocked(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	if logger.file != nil {
		logger.file.Close()
	}

	logger.file = file
	log.Println("File logging enabled:", filePath)
	return nil
}

func EnableDailyRotation(dirPath string, prefix string) error {
	mu.Lock()
	defer mu.Unlock()

	logger.logDir = dirPath
	logger.prefix = prefix
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	return rotateLogFileLocked()
}

func rotateLogFileLocked() error {
	now := time.Now()
	dateStr := now.Format("2006-01-02")
	logger.lastDate = dateStr

	fileName := fmt.Sprintf("%s.log", dateStr)
	if logger.prefix != "" {
		fileName = fmt.Sprintf("%s-%s.log", logger.prefix, dateStr)
	}

	logPath := fmt.Sprintf("%s/%s", strings.TrimSuffix(logger.logDir, "/"), fileName)
	return enableFileLoggingLocked(logPath)
}

func DisableFileLogging() {
	mu.Lock()
	defer mu.Unlock()

	if logger.file != nil {
		logger.file.Close()
		logger.file = nil
		log.Println("File logging disabled")
	}
}

func logMessage(level LogLevel, component string, message string, fields map[string]interface{}) {
	if level < currentLevel {
		return
	}

	entry := LogEntry{
		Level:     logLevelNames[level],
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Component: component,
		Message:   message,
		Fields:    fields,
	}

	if pc, file, line, ok := runtime.Caller(2); ok {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			entry.Caller = fmt.Sprintf("%s:%d (%s)", file, line, fn.Name())
		}
	}

	if logger.file != nil {
		// Check for date rotation
		if logger.logDir != "" {
			currentDate := time.Now().Format("2006-01-02")
			if currentDate != logger.lastDate {
				rotateLogFileLocked()
			}
		}

		jsonData, err := json.Marshal(entry)
		if err == nil {
			logger.file.WriteString(string(jsonData) + "\n")
		}
	}

	var fieldStr string
	if len(fields) > 0 {
		fieldStr = " " + formatFields(fields)
	}

	logLine := fmt.Sprintf("[%s] [%s]%s %s%s",
		entry.Timestamp,
		logLevelNames[level],
		formatComponent(component),
		message,
		fieldStr,
	)

	log.Println(logLine)

	if level == FATAL {
		os.Exit(1)
	}
}

func formatComponent(component string) string {
	if component == "" {
		return ""
	}
	return fmt.Sprintf(" %s:", component)
}

func formatFields(fields map[string]interface{}) string {
	var parts []string
	for k, v := range fields {
		parts = append(parts, fmt.Sprintf("%s=%v", k, v))
	}
	return fmt.Sprintf("{%s}", strings.Join(parts, ", "))
}

func Debug(message string) {
	logMessage(DEBUG, "", message, nil)
}

func DebugC(component string, message string) {
	logMessage(DEBUG, component, message, nil)
}

func DebugF(message string, fields map[string]interface{}) {
	logMessage(DEBUG, "", message, fields)
}

func DebugCF(component string, message string, fields map[string]interface{}) {
	logMessage(DEBUG, component, message, fields)
}

func Info(message string) {
	logMessage(INFO, "", message, nil)
}

func InfoC(component string, message string) {
	logMessage(INFO, component, message, nil)
}

func InfoF(message string, fields map[string]interface{}) {
	logMessage(INFO, "", message, fields)
}

func InfoCF(component string, message string, fields map[string]interface{}) {
	logMessage(INFO, component, message, fields)
}

func Warn(message string) {
	logMessage(WARN, "", message, nil)
}

func WarnC(component string, message string) {
	logMessage(WARN, component, message, nil)
}

func WarnF(message string, fields map[string]interface{}) {
	logMessage(WARN, "", message, fields)
}

func WarnCF(component string, message string, fields map[string]interface{}) {
	logMessage(WARN, component, message, fields)
}

func Error(message string) {
	logMessage(ERROR, "", message, nil)
}

func ErrorC(component string, message string) {
	logMessage(ERROR, component, message, nil)
}

func ErrorF(message string, fields map[string]interface{}) {
	logMessage(ERROR, "", message, fields)
}

func ErrorCF(component string, message string, fields map[string]interface{}) {
	logMessage(ERROR, component, message, fields)
}

func Fatal(message string) {
	logMessage(FATAL, "", message, nil)
}

func FatalC(component string, message string) {
	logMessage(FATAL, component, message, nil)
}

func FatalF(message string, fields map[string]interface{}) {
	logMessage(FATAL, "", message, fields)
}

func FatalCF(component string, message string, fields map[string]interface{}) {
	logMessage(FATAL, component, message, fields)
}
