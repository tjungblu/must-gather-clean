package obfuscator

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/openshift/must-gather-clean/pkg/schema"
)

const (
	obfuscatedStaticIPv4 = "xxx.xxx.xxx.xxx"
	// 4.228.250.625 possible addresses can fit into 10 chars
	consistentIPv4Template = "x-ipv4-%010d-x"
)

var (
	ipv4Pattern   = regexp.MustCompile(`\b(([0-9]{1,3}[.]){3}|([0-9]{1,3}[-]){3})([0-9]{1,3})`)
	excludedIPv4s = map[string]struct{}{
		"127.0.0.1": {},
		"0.0.0.0":   {},
	}
)

type ipv4Obfuscator struct {
	ReplacementTracker
	generator       *generator
	replacementType schema.ObfuscateReplacementType
}

func (o *ipv4Obfuscator) Path(s string) string {
	return o.replace(s)
}

func (o *ipv4Obfuscator) Contents(s string) string {
	return o.replace(s)
}

func (o *ipv4Obfuscator) replace(s string) string {
	output := s
	ipMatches := ipv4Pattern.FindAllString(output, -1)

	for _, m := range ipMatches {
		// if the match is in the exclude-list then do not replace.
		if _, ok := excludedIPv4s[m]; ok {
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

func NewIPv4Obfuscator(replacementType schema.ObfuscateReplacementType) (Obfuscator, error) {
	if replacementType != schema.ObfuscateReplacementTypeStatic && replacementType != schema.ObfuscateReplacementTypeConsistent {
		return nil, fmt.Errorf("unsupported replacement type: %s", replacementType)
	}
	return &ipv4Obfuscator{
		ReplacementTracker: NewSimpleTracker(),
		generator:          newGenerator(consistentIPv4Template, obfuscatedStaticIPv4, 9999999999),
		replacementType:    replacementType,
	}, nil
}
