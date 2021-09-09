package obfuscator

import (
	"fmt"
	"strings"

	"github.com/openshift/must-gather-clean/pkg/match"
	"github.com/openshift/must-gather-clean/pkg/schema"
)

type fastIpv4Obfuscator struct {
	ReplacementTracker
	generator       *generator
	replacementType schema.ObfuscateReplacementType
}

func (o *fastIpv4Obfuscator) Path(s string) string {
	return o.replace(s)
}

func (o *fastIpv4Obfuscator) Contents(s string) string {
	return o.replace(s)
}

func (o *fastIpv4Obfuscator) replace(s string) string {
	sb := strings.Builder{}
	lastWriteIndex := 0
	for i := 0; i < len(s); i++ {
		ip, start, end := match.FindNextIPv4(s[i:])
		if ip != "" {
			var replacement string
			switch o.replacementType {
			case schema.ObfuscateReplacementTypeStatic:
				replacement = o.GenerateIfAbsent(ip, o.generator.generateStaticReplacement)
			case schema.ObfuscateReplacementTypeConsistent:
				replacement = o.GenerateIfAbsent(ip, o.generator.generateConsistentReplacement)
			}

			o.ReplacementTracker.AddReplacement(ip, replacement)
			// add the original format into the replacement tracker too
			o.ReplacementTracker.AddReplacement(s[i:][start:end], replacement)

			sb.WriteString(s[lastWriteIndex : i+start])
			sb.WriteString(replacement)
			lastWriteIndex = i + end
			i = lastWriteIndex - 1 // since it will increment it again in the loop
		} else {
			sb.WriteString(string(s[i]))
			lastWriteIndex = i
		}
	}

	if sb.Len() == 0 {
		return s
	}

	return sb.String()
}

func NewFastIPv4Obfuscator(replacementType schema.ObfuscateReplacementType) (Obfuscator, error) {
	if replacementType != schema.ObfuscateReplacementTypeStatic && replacementType != schema.ObfuscateReplacementTypeConsistent {
		return nil, fmt.Errorf("unsupported replacement type: %s", replacementType)
	}
	return &fastIpv4Obfuscator{
		ReplacementTracker: NewSimpleTracker(),
		generator:          newGenerator(consistentIPv4Template, obfuscatedStaticIPv4, 9999999999),
		replacementType:    replacementType,
	}, nil
}
