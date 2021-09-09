package obfuscator

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/openshift/must-gather-clean/pkg/schema"
)

const (
	obfuscatedStaticIPv6   = "xxxx:xxxx:xxxx:xxxx:xxxx:xxxx:xxxx:xxxx"
	consistentIPv6Template = "xx-ipv6-%018d-xx"
)

var (
	// ipv6re is not perfect. it can still catch words like :face:bad as a valid ipv6 address
	ipv6Pattern   = regexp.MustCompile(`([a-f0-9]{0,4}[:]){1,8}[a-f0-9]{1,4}`)
	excludedIPv6s = map[string]struct{}{
		"::1": {},
	}
)

type ipv6Obfuscator struct {
	ReplacementTracker
	generator       *generator
	replacementType schema.ObfuscateReplacementType
}

func (o *ipv6Obfuscator) Path(s string) string {
	return o.replace(s)
}

func (o *ipv6Obfuscator) Contents(s string) string {
	return o.replace(s)
}

func (o *ipv6Obfuscator) replace(s string) string {
	output := s
	ipMatches := ipv6Pattern.FindAllString(output, -1)

	for _, m := range ipMatches {
		// if the match is in the exclude-list then do not replace.
		if _, ok := excludedIPv6s[m]; ok {
			continue
		}

		cleaned := strings.ReplaceAll(m, "-", ".")
		if ip := net.ParseIP(cleaned); ip != nil {
			var replacement string
			switch o.replacementType {
			case schema.ObfuscateReplacementTypeStatic:
				replacement = o.GenerateIfAbsent(cleaned, o.generator.generateStaticReplacement)
			case schema.ObfuscateReplacementTypeConsistent:
				replacement = o.GenerateIfAbsent(cleaned, o.generator.generateConsistentReplacement)
			}
			output = strings.ReplaceAll(output, m, replacement)
			o.ReplacementTracker.AddReplacement(m, replacement)
		}
	}
	return output
}

func NewIPv6Obfuscator(replacementType schema.ObfuscateReplacementType) (Obfuscator, error) {
	if replacementType != schema.ObfuscateReplacementTypeStatic && replacementType != schema.ObfuscateReplacementTypeConsistent {
		return nil, fmt.Errorf("unsupported replacement type: %s", replacementType)
	}
	return &ipv6Obfuscator{
		ReplacementTracker: NewSimpleTracker(),
		generator:          newGenerator(consistentIPv6Template, obfuscatedStaticIPv6, 999999999999999999),
		replacementType:    replacementType,
	}, nil
}
