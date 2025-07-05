package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/gkwa/belgianlake/internal/logger"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestCustomLogger(t *testing.T) {
	var buf bytes.Buffer

	zapConfig := zap.NewDevelopmentConfig()
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zapLogger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zapConfig.EncoderConfig),
		zapcore.AddSync(&buf),
		zapcore.DebugLevel,
	))

	customLogger := zapr.NewLogger(zapLogger)

	cliLogger = customLogger

	cmd := rootCmd
	cmd.SetArgs([]string{"version"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	logOutput := buf.String()
	if logOutput == "" {
		t.Error("Expected log output, but got none")
	}

	// Check if the log contains version information
	if !strings.Contains(logOutput, "Version information") {
		t.Error("Expected version information in log output")
	}

	t.Logf("Log output: %s", logOutput)
}

func TestJSONLogger(t *testing.T) {
	oldVerbose, oldLogFormat := verbose, logFormat
	verbose, logFormat = true, "json"
	defer func() {
		verbose, logFormat = oldVerbose, oldLogFormat
	}()

	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	customLogger := logger.NewConsoleLogger(verbose, logFormat == "json")
	cliLogger = customLogger

	cmd := rootCmd
	cmd.SetArgs([]string{"version"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		t.Fatalf("Failed to copy log output: %v", err)
	}
	logOutput := buf.String()

	if logOutput == "" {
		t.Error("Expected log output, but got none")
	}

	lines := strings.Split(strings.TrimSpace(logOutput), "\n")
	validJSONFound := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var jsonMap map[string]interface{}
		err := json.Unmarshal([]byte(line), &jsonMap)
		if err != nil {
			t.Errorf("Expected valid JSON, but got error: %v for line: %s", err, line)
		} else {
			validJSONFound = true
		}
	}

	if !validJSONFound {
		t.Error("No valid JSON log entries found")
	}

	t.Logf("Log output: %s", logOutput)
}
