package inventory

import (
	"bufio"
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewCSVHasHeaders(t *testing.T) {
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	c, err := NewCSV(writer)
	if err != nil {
		t.Fatal(err)
	}
	c.Flush()

	expected := strings.Join(csvHeaders, ",")
	actual := strings.TrimSpace(buf.String())

	require.Equal(t, expected, actual, "wrote unexpected csv headers")
}

func TestNewCSVWritesRow(t *testing.T) {
	var buf bytes.Buffer
	writer := bufio.NewWriter(&buf)

	c, err := NewCSV(writer)
	if err != nil {
		t.Fatal(err)
	}

	require.NoError(t, c.WriteRow(testRow))
	c.Flush()

	expected := strings.Join(testRow.StringSlice(), ",")
	actual := strings.TrimSpace(buf.String())

	require.Contains(t, actual, expected, "failed to find row")
}
