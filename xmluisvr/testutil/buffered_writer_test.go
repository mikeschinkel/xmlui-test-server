package testutil

import (
	"errors"
	"strings"
	"testing"

	"github.com/xmlui-org/xmlui-test-server/xmluisvr/cliutil"
)

func TestBufferedWriter_Basic(t *testing.T) {
	writer := NewBufferedWriter()

	// Test basic Printf
	writer.Printf("Hello %s", "World")
	if !writer.ContainsStdout("Hello World") {
		t.Errorf("Expected stdout to contain 'Hello World', got: %q", writer.GetStdout())
	}

	// Test basic Errorf
	writer.Errorf("Error: %s", "something failed")
	if !writer.ContainsStderr("Error: something failed") {
		t.Errorf("Expected doterr to contain 'Error: something failed', got: %q", writer.GetStderr())
	}
}

func TestBufferedWriter_Interface(t *testing.T) {
	var w cliutil.Writer = NewBufferedWriter()

	// Test that all interface methods are available
	w.Printf("test")
	w.Errorf("error test")
	loudWriter := w.Loud()
	v2Writer := w.V2()
	v3Writer := w.V3()

	if loudWriter == nil {
		t.Error("Loud() should return a non-nil Writer")
	}
	if v2Writer == nil {
		t.Error("V2() should return a non-nil Writer")
	}
	if v3Writer == nil {
		t.Error("V3() should return a non-nil Writer")
	}
}

func TestBufferedWriter_VerbosityLevels(t *testing.T) {
	writer := NewBufferedWriterWithVerbosity(2)

	// Level 1 should work (verbosity 2 >= useLevel 1)
	writer.Printf("Level 1 message")
	if !writer.ContainsStdout("Level 1 message") {
		t.Error("Level 1 message should be captured")
	}

	writer.Reset()

	// Level 2 should work (verbosity 2 >= useLevel 2)
	v2 := writer.V2()
	v2.Printf("Level 2 message")
	if !writer.ContainsStdout("Level 2 message") {
		t.Error("Level 2 message should be captured")
	}

	writer.Reset()

	// Level 3 should NOT work (verbosity 2 < useLevel 3)
	v3 := writer.V3()
	v3.Printf("Level 3 message")
	if writer.ContainsStdout("Level 3 message") {
		t.Error("Level 3 message should NOT be captured with verbosity 2")
	}
}

func TestBufferedWriter_QuietMode(t *testing.T) {
	writer := NewBufferedWriter()
	writer.SetQuiet(true)

	// Printf should be suppressed in quiet mode
	writer.Printf("Should not appear")
	if writer.ContainsStdout("Should not appear") {
		t.Error("Printf should be suppressed in quiet mode")
	}

	// Errorf should still work in quiet mode
	writer.Errorf("Error should appear")
	if !writer.ContainsStderr("Error should appear") {
		t.Error("Errorf should work even in quiet mode")
	}

	// Loud() should bypass quiet mode
	loud := writer.Loud()
	loud.Printf("Loud message")
	if !writer.ContainsStdout("Loud message") {
		t.Error("Loud() should bypass quiet mode")
	}
}

func TestBufferedWriter_ErrorFormatting(t *testing.T) {
	writer := NewBufferedWriter()

	// Test error with newlines gets flattened
	err := errors.New("line 1\nline 2\nline 3")
	writer.Errorf("Error occurred: %v", err)

	stderr := writer.GetStderr()
	if !strings.Contains(stderr, "line 1; line 2; line 3") {
		t.Errorf("Expected error newlines to be replaced with semicolons, got: %q", stderr)
	}
}

func TestBufferedWriter_HelperMethods(t *testing.T) {
	writer := NewBufferedWriter()

	// Test line counting
	writer.Printf("Line 1\n")
	writer.Printf("Line 2\n")
	writer.Printf("\n") // Empty line should be ignored
	writer.Printf("Line 3\n")

	lines := writer.GetStdoutLines()
	if len(lines) != 3 {
		t.Errorf("Expected 3 non-empty lines, got %d: %v", len(lines), lines)
	}

	if writer.CountStdoutLines() != 3 {
		t.Errorf("Expected CountStdoutLines() to return 3, got %d", writer.CountStdoutLines())
	}

	// Test reset
	writer.Reset()
	if writer.GetStdout() != "" {
		t.Error("Expected stdout to be empty after Reset()")
	}
	if writer.GetStderr() != "" {
		t.Error("Expected doterr to be empty after Reset()")
	}
}

func TestBufferedWriter_SharedBuffers(t *testing.T) {
	writer := NewBufferedWriter()

	// Test that V2 and V3 share the same buffers
	writer.Printf("Main message\n")
	writer.V2().Printf("V2 message\n")
	writer.V3().Printf("V3 message\n")

	stdout := writer.GetStdout()
	if !strings.Contains(stdout, "Main message") {
		t.Error("Main message should be in stdout")
	}
	if !strings.Contains(stdout, "V2 message") {
		t.Error("V2 message should be in shared stdout buffer")
	}
	if !strings.Contains(stdout, "V3 message") {
		t.Error("V3 message should be in shared stdout buffer")
	}
}

func TestBufferedWriter_ConcurrentAccess(t *testing.T) {
	writer := NewBufferedWriter()

	// Test concurrent writes don't cause data races
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			writer.Printf("goroutine1-%d\n", i)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			writer.Errorf("goroutine2-%d\n", i)
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Just verify we don't crash and have some content
	if writer.CountStdoutLines() == 0 {
		t.Error("Expected some stdout lines from concurrent writes")
	}
	if writer.CountStderrLines() == 0 {
		t.Error("Expected some doterr lines from concurrent writes")
	}
}
