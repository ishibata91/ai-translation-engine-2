package errorwrapcases

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

var ErrConverted = errors.New("converted")
var _ = localHelper

func downstream() error {
	return errors.New("boom")
}

func ExportedReturnErr() error {
	err := downstream()
	if err != nil {
		return err // want "error wrap: wrap returned error with context before crossing a package boundary"
	}
	return nil
}

func ExportedFmtNoWrap() error {
	err := downstream()
	if err != nil {
		return fmt.Errorf("downstream failed: %v", err) // want "error wrap: use %w when returning fmt.Errorf with an underlying error"
	}
	return nil
}

func IgnoreAssignedError() {
	_, _ = os.Open("missing.txt") // want "error wrap: do not ignore errors outside cleanup or best-effort paths"
}

func IgnoreExpressionError() {
	os.Remove("missing.txt") // want "error wrap: do not ignore errors outside cleanup or best-effort paths"
}

func AllowedCleanup() error {
	file, err := os.CreateTemp("", "wrapcheck")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}

	defer func() {
		_ = file.Close()
	}()

	return nil
}

func AllowedConvertedError() error {
	err := downstream()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrConverted
		}
		return ErrConverted
	}
	return nil
}

func RejectSelfWrapReassignment() error {
	err := downstream()
	if err != nil {
		err = fmt.Errorf("retry failed: %w", err) // want "error wrap: do not reassign an error to fmt.Errorf\\(... %w, err\\) on the same variable"
		return err
	}
	return nil
}

func AllowedWrapIntoNewVariable() error {
	err := downstream()
	if err != nil {
		wrappedErr := fmt.Errorf("retry failed: %w", err)
		return wrappedErr
	}
	return nil
}

func AllowedDiscardNonErrorResult() error {
	_, err := os.CreateTemp("", "discard")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	return nil
}

func AllowedBestEffortOutput() {
	fmt.Println("best effort output")
}

func AllowedBuilderWrite() string {
	var sb strings.Builder
	sb.WriteString("builder output")
	return sb.String()
}

func localHelper() error {
	err := downstream()
	if err != nil {
		return err
	}
	return nil
}
