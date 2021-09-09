package match

import (
	"net"
)

const maxOctetSize = 0xFF

// FindNextIPv4 tries to find the next IPv4 address in the given string s.
// Returns the normalized found IP address, the starting index (inclusive) and the end index (exclusive)
func FindNextIPv4(s string) (string, int, int) {
	i := 0
	for ; i < len(s); i++ {
		if s[i] >= '1' && s[i] <= '9' {
			foundIp, nextIndex := tryParseIPv4(s[i:])
			if foundIp != nil {
				return foundIp.String(), i, i + nextIndex
			}

			i = i + nextIndex
		}
	}

	return "", 0, i
}

func tryParseIPv4(s string) (net.IP, int) {
	oLen := len(s)
	var lastSep uint8
	var p [net.IPv4len]byte
	for i := 0; i < net.IPv4len; i++ {
		if len(s) == 0 {
			// Missing octets.
			return nil, oLen - len(s)
		}
		if i > 0 {
			// this ensures we can parse consistent usage of dots and dashes, and no weird mix forms
			if (s[0] != '.' && s[0] != '-') || (lastSep != 0 && s[0] != lastSep) {
				return nil, oLen - len(s)
			}
			lastSep = s[0]
			s = s[1:]
		}
		n, c, ok := parseOctet(s)
		if !ok || n > 0xFF {
			return nil, oLen - len(s)
		}
		s = s[c:]
		p[i] = byte(n)
	}

	return net.IPv4(p[0], p[1], p[2], p[3]), oLen - len(s)
}

// Returns number, characters consumed, success.
func parseOctet(s string) (n int, i int, ok bool) {
	n = 0
	for i = 0; i < len(s) && '0' <= s[i] && s[i] <= '9'; i++ {
		n = n*10 + int(s[i]-'0')
		if n >= maxOctetSize {
			return maxOctetSize, i, false
		}
	}
	if i == 0 {
		return 0, 0, false
	}
	return n, i, true
}
