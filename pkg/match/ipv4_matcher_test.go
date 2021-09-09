package match

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchingIPv4s(t *testing.T) {
	for _, tc := range []struct {
		name               string
		input              string
		expectedIP         string
		expectedStartIndex int
		expectedEndIndex   int
		report             map[string]string
	}{
		{
			name:               "valid ipv4 address",
			input:              "received request from 192.168.1.10",
			expectedIP:         "192.168.1.10",
			expectedStartIndex: 22,
			expectedEndIndex:   34,
		},
		{
			name:               "valid ipv4 address beginning",
			input:              "192.168.1.10 received request",
			expectedIP:         "192.168.1.10",
			expectedStartIndex: 0,
			expectedEndIndex:   12,
		},
		{
			name:               "valid ipv4 address only",
			input:              "192.168.1.10",
			expectedIP:         "192.168.1.10",
			expectedStartIndex: 0,
			expectedEndIndex:   12,
		},
		{
			name:               "ipv4 in words",
			input:              "calling https://192.168.1.20/metrics for values",
			expectedIP:         "192.168.1.20",
			expectedStartIndex: 16,
			expectedEndIndex:   28,
		},
		{
			name:               "multiple ipv4s finds first",
			input:              "received request from 192.168.1.20 proxied through 192.168.1.3",
			expectedIP:         "192.168.1.20",
			expectedStartIndex: 22,
			expectedEndIndex:   34,
		},
		{
			name:               "parse google dns",
			input:              "dns forwarded to 8.8.8.8 and was OK",
			expectedIP:         "8.8.8.8",
			expectedStartIndex: 17,
			expectedEndIndex:   24,
		},
		{
			name:               "non standard ipv4",
			input:              "ip-10-0-129-220.ec2.aws.yaml",
			expectedIP:         "10.0.129.220",
			expectedStartIndex: 3,
			expectedEndIndex:   15,
		},
		{
			name:               "non-standard ipv4 with bad separator",
			input:              "ip+10+0+129+220.ec2.aws.yaml",
			expectedIP:         "",
			expectedStartIndex: 0,
			expectedEndIndex:   28,
		},
		{
			name:               "standard ipv4 and standard ipv4",
			input:              "obfuscate 10.0.129.220 and 10-0-129-220",
			expectedIP:         "10.0.129.220",
			expectedStartIndex: 10,
			expectedEndIndex:   22,
		},
		{
			name:               "OCP nightly version false positive",
			input:              "version: 4.8.0-0.nightly-2021-07-31-065602",
			expectedIP:         "",
			expectedStartIndex: 0,
			expectedEndIndex:   43,
		},
		{
			name:               "OCP version x.y.z",
			input:              "version: 4.8.12",
			expectedIP:         "",
			expectedStartIndex: 0,
			expectedEndIndex:   16,
		},
		{
			name:               "excluded ipv4 address",
			input:              "Listening on 0.0.0.0:8080",
			expectedIP:         "",
			expectedStartIndex: 0,
			expectedEndIndex:   26,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			nextIPv4, start, end := FindNextIPv4(tc.input)
			assert.Equal(t, tc.expectedIP, nextIPv4)
			assert.Equal(t, tc.expectedStartIndex, start)
			assert.Equal(t, tc.expectedEndIndex, end)
		})
	}
}
