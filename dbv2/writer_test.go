package dbv2

import (
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreateFileAndWriteIt(t *testing.T) {
	contents := []struct {
		timestamp, seriesID uint64
		values              []string
	}{
		{
			timestamp: 1,
			seriesID:  1,
			values:    []string{"234.5", "11.2", "3"},
		},
		{
			timestamp: 10,
			seriesID:  1,
			values:    []string{"234.5", "11.2", "3"},
		},
		{
			timestamp: 100,
			seriesID:  1,
			values:    []string{"234.5", "11.2", "3"},
		},
		{
			timestamp: 200,
			seriesID:  1,
			values:    []string{"234.5", "11.2", "3"},
		},
		{
			timestamp: 300,
			seriesID:  1,
			values:    []string{"234.5", "11.2", "3"},
		},
	}
	expected := `maxValidLength: 100 chars

1|1|234.5:11.2:3\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00
10|1|234.5:11.2:3\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00
100|1|234.5:11.2:3\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00
200|1|234.5:11.2:3\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00
300|1|234.5:11.2:3\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00`
	testFile := "test_writer_file"
	dtbl, err := CreateDataTable(testFile)
	require.NoError(t, err)

	tbBuffer := NewTableBuffer(dtbl, &sync.RWMutex{}, 2, 1000, time.Minute)
	for _, c := range contents {
		err := tbBuffer.Write(c.timestamp, c.seriesID, ConvertValueToValueSet(c.values[0], c.values[1], c.values[2]))
		require.NoError(t, err, "writing contents")
		// Keeping flushToIOBuffer(true) in for loop, is the actual tough part for the test to pass. Passing here would mean there is no
		// duplication of data (which is what we want). Ideally, this would be kept after the loop for write is done, but since this is testing,
		// we want to make sure for edge cases.
		err = tbBuffer.flushToIOBuffer(true)
		require.NoError(t, err)
	}
	bSlice, err := ioutil.ReadFile(testFile)
	require.NoError(t, err)
	dr, err := NewDataReader(testFile)
	require.NoError(t, err)
	err = dr.Parse()
	require.NoError(t, err)
	require.Equal(t, []byte(expected), bSlice, "matching write result")
	require.NoError(t, os.Remove(testFile))
}

func TestUnorderedInserts(t *testing.T) {
	contents := []struct {
		timestamp, seriesID uint64
		values              []string
	}{
		{
			timestamp: 1000,
			seriesID:  1,
			values:    []string{"234.5", "11.2", "3"},
		},
		{
			timestamp: 10,
			seriesID:  1,
			values:    []string{"234.5", "11.2", "3"},
		},
		{
			timestamp: 500,
			seriesID:  1,
			values:    []string{"234.5", "11.2", "3"},
		},
		{
			timestamp: 200,
			seriesID:  1,
			values:    []string{"234.5", "11.2", "3"},
		},
		{
			timestamp: 300000,
			seriesID:  1,
			values:    []string{"234.5", "11.2", "3"},
		},
		{
			timestamp: 700,
			seriesID:  1,
			values:    []string{"234.5", "11.2", "3"},
		},
		{
			timestamp: 1000,
			seriesID:  1,
			values:    []string{"234.5", "11.2", "3"},
		},
		{
			timestamp: 5001,
			seriesID:  1,
			values:    []string{"234.5", "11.2", "3"},
		},
		{
			timestamp: 20,
			seriesID:  1,
			values:    []string{"234.5", "11.2", "3"},
		},
		{
			timestamp: 3000,
			seriesID:  1,
			values:    []string{"234.5", "11.2", "3"},
		},
	}
	expected := `maxValidLength: 100 chars

10|1|234.5:11.2:3
20|1|234.5:11.2:3
200|1|234.5:11.2:3
500|1|234.5:11.2:3
700|1|234.5:11.2:3
1000|1|234.5:11.2:3
1000|1|234.5:11.2:3
3000|1|234.5:11.2:3
5001|1|234.5:11.2:3
300000|1|234.5:11.2:3
`
	testFile := "test_unordered_inserts_file"
	dtbl, err := CreateDataTable(testFile)
	require.NoError(t, err)
	tbBuffer := NewTableBuffer(dtbl, &sync.RWMutex{}, uint64(len(contents)), 1000, time.Minute)
	for _, c := range contents {
		err := tbBuffer.Write(c.timestamp, c.seriesID, ConvertValueToValueSet(c.values[0], c.values[1], c.values[2]))
		require.NoError(t, err, "writing contents")
	}
	err = tbBuffer.flushToIOBuffer(true)
	require.NoError(t, err)
	bSlice, err := ioutil.ReadFile(testFile)
	require.NoError(t, err)
	require.Equal(t, []byte(expected), bSlice, "matching write result")
	require.NoError(t, os.Remove(testFile))
}