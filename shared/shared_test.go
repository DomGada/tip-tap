package shared

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

// TestSizeof verifies Sizeof returns the correct value for known bit widths.
// Sizeof starts with num=1, toAdd=1, and each iteration does num = (num<<1) + 1,
// producing 2^n - 1 + 1 = 2^(n+1) - 1... tracing manually:
//   bits=0 → 1
//   bits=1 → 3
//   bits=2 → 7
//   bits=8 → 511
//   bits=16 → 131071
func TestSizeof(t *testing.T) {
	cases := []struct {
		bits     int
		expected int
	}{
		{bits: 0, expected: 1},
		{bits: 1, expected: 3},
		{bits: 2, expected: 7},
		{bits: 8, expected: 511},
		{bits: 16, expected: 131071},
	}

	for _, tc := range cases {
		got := Sizeof(tc.bits)
		if got != tc.expected {
			t.Errorf("Sizeof(%d) = %d, want %d", tc.bits, got, tc.expected)
		}
	}
}

// TestLoggerLoginfo verifies that Loginfo does not panic and produces output.
func TestLoggerLoginfo(t *testing.T) {
	var buf bytes.Buffer
	original := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(original)

	var logger Logger
	logger.Loginfo("hello from loginfo")

	output := buf.String()
	if !strings.Contains(output, "hello from loginfo") {
		t.Errorf("Loginfo output %q does not contain expected message", output)
	}
}

// TestLoggerLogerr verifies that Logerr does not panic and writes to stderr.
func TestLoggerLogerr(t *testing.T) {
	// Capture stderr via os.Pipe so we can confirm Logerr writes there.
	origStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stderr = w

	var logger Logger
	logger.Logerr("hello from logerr")

	// Close the write end so reading does not block.
	w.Close()
	os.Stderr = origStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	r.Close()

	output := buf.String()
	if !strings.Contains(output, "hello from logerr") {
		t.Errorf("Logerr output %q does not contain expected message", output)
	}
}

// TestLoggerLazyInit verifies that a zero-value Logger initializes its internal
// loggers on first use without panicking.
func TestLoggerLazyInit(t *testing.T) {
	var logger Logger

	if logger._stdout != nil {
		t.Error("expected _stdout to be nil before first use")
	}
	if logger._stderr != nil {
		t.Error("expected _stderr to be nil before first use")
	}

	// Redirect default log output so the test does not pollute stdout.
	var buf bytes.Buffer
	original := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(original)

	// First call should trigger lazy initialization without panicking.
	logger.Loginfo("init test")

	if logger._stdout == nil {
		t.Error("expected _stdout to be non-nil after first Loginfo call")
	}

	// Redirect stderr for Logerr init check.
	origStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stderr = w

	logger.Logerr("init err test")

	w.Close()
	os.Stderr = origStderr
	r.Close()

	if logger._stderr == nil {
		t.Error("expected _stderr to be non-nil after first Logerr call")
	}
}

// TestErrorCodes verifies ERR_OPENER_BAD_CHOICE equals 11.
// iota=0 for blank _, iota=1 for ERR_OPENER_BAD_CHOICE, so value = 1 + 10 = 11.
func TestErrorCodes(t *testing.T) {
	const expected = 11
	if ERR_OPENER_BAD_CHOICE != expected {
		t.Errorf("ERR_OPENER_BAD_CHOICE = %d, want %d", ERR_OPENER_BAD_CHOICE, expected)
	}
}
