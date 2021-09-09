package obfuscator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/openshift/must-gather-clean/pkg/schema"
)

func TestIPv6ObfuscatorStatic(t *testing.T) {
	for _, tc := range []struct {
		name   string
		input  string
		output string
		report map[string]string
	}{
		{
			name:   "valid ipv6 address",
			input:  "received request from 2001:db8::ff00:42:8329",
			output: "received request from xxxx:xxxx:xxxx:xxxx:xxxx:xxxx:xxxx:xxxx",
			report: map[string]string{
				"2001:db8::ff00:42:8329": obfuscatedStaticIPv6,
			},
		},
		{
			name:   "excluded ipv6 address",
			input:  "Listening on [::1]:8080",
			output: "Listening on [::1]:8080",
			report: map[string]string{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o, err := NewIPv6Obfuscator(schema.ObfuscateReplacementTypeStatic)
			require.NoError(t, err)
			output := o.Contents(tc.input)
			assert.Equal(t, tc.output, output)
			assert.Equal(t, tc.report, o.Report())
		})
	}
}

func TestIPv6ObfuscatorConsistent(t *testing.T) {
	for _, tc := range []struct {
		name   string
		input  []string
		output []string
		report map[string]string
	}{
		{
			name:   "valid ipv6 address",
			input:  []string{"received request from 2001:db8::ff00:42:8329"},
			output: []string{"received request from xx-ipv6-000000000000000001-xx"},
			report: map[string]string{
				"2001:db8::ff00:42:8329": "xx-ipv6-000000000000000001-xx",
			},
		},
		{
			name:   "mixed ipv4 and ipv6",
			input:  []string{"tunneling ::2fa:bf9 as 192.168.1.30"},
			output: []string{"tunneling xx-ipv6-000000000000000001-xx as 192.168.1.30"},
			report: map[string]string{
				"::2fa:bf9": "xx-ipv6-000000000000000001-xx",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o, err := NewIPv6Obfuscator(schema.ObfuscateReplacementTypeConsistent)
			require.NoError(t, err)
			for i := 0; i < len(tc.input); i++ {
				assert.Equal(t, tc.output[i], o.Contents(tc.input[i]))
			}
			assert.Equal(t, tc.report, o.Report())
		})
	}
}
