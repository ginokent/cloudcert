// nolint: testpackage
package repository

import (
	"io"
	"testing"
)

func TestLegoStdLogger(t *testing.T) {
	t.Parallel()

	testLegoLogger := newLegoStdLogger(io.Discard)

	testLegoLogger.SetInternalLogger(testLegoLogger.GetInternalLogger().RenewWriter(io.Discard))

	t.Run("success()", func(t *testing.T) {
		t.Parallel()

		// Print
		testLegoLogger.Print("[INFO]")
		testLegoLogger.Print("[WARN]")
		testLegoLogger.Print("")
		// Println
		testLegoLogger.Println("[INFO]")
		testLegoLogger.Println("[WARN]")
		testLegoLogger.Println("")
		// Printf
		testLegoLogger.Printf("[INFO]")
		testLegoLogger.Printf("[WARN]")
		testLegoLogger.Printf("")
		// Fatal
		func() {
			defer func() { _ = recover() }()
			testLegoLogger.Fatal("")
		}()
		// Fatalln
		func() {
			defer func() { _ = recover() }()
			testLegoLogger.Fatalln("")
		}()
		// Fatalf
		func() {
			defer func() { _ = recover() }()
			testLegoLogger.Fatalf("")
		}()
	})
}
