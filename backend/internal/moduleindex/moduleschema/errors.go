package moduleschema

import (
	"regexp"
)

var stripColorRe = regexp.MustCompile("\x1b\\[(.*?)m")

type SchemaExtractionFailedError struct {
	Output []byte
	Cause  error
}

func (s SchemaExtractionFailedError) Error() string {
	if len(s.Output) > 0 {
		return "Schema extraction failed: " + s.Cause.Error() + " (tofu output: " + s.OutputString() + ")"
	}
	return "Schema extraction failed: " + s.Cause.Error()
}

func (s SchemaExtractionFailedError) Unwrap() error {
	return s.Cause
}

func (s SchemaExtractionFailedError) OutputString() string {
	outputString := string(s.Output)
	outputString = stripColorRe.ReplaceAllString(outputString, "")
	return outputString
}

func (s SchemaExtractionFailedError) RawOutput() []byte {
	return s.Output
}
