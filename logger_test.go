package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHere(t *testing.T) {
	require := require.New(t)

	l := sampleFunc()
	require.Equal("github.com/dc0d/logger.sampleFunc", l.FuncName)
	require.Contains(l.FileName, "/github.com/dc0d/logger/template_test.go")
	require.Equal(4, l.FileLine)
	require.Equal("logger/template_test.go@4:github.com/dc0d/logger.sampleFunc()", l.long)
	require.Equal("logger/template_test.go@4:sampleFunc()", l.short)
}
