package obfuscator

import (
	"hash/fnv"
	"runtime"
	"sync"

	"k8s.io/klog/v2"
)

type GenerateReplacement func(string) string

// ReplacementTracker is used to track and generate replacements used by obfuscators
type ReplacementTracker interface {
	// Initialize initializes the tracker with some existing replacements. It should be called only once and before
	// the first use of GetReplacement or AddReplacement
	Initialize(replacements map[string]string)

	// Report returns a mapping of strings which were replaced.
	Report() map[string]string

	// AddReplacement will add a replacement along with its original string to the report.
	// If there is an existing value that does not match the given replacement, it will exit with a non-zero status.
	AddReplacement(original string, replacement string)

	// GenerateIfAbsent returns the previously used replacement if already set. If the replacement is not present then it
	// uses the GenerateReplacement function to generate a replacement. Generator should not be empty. The original
	// parameter must be used for lookup and the key parameter to generate the replacement.
	GenerateIfAbsent(original string, key string, generator GenerateReplacement) string
}

type SimpleTracker struct {
	lock    sync.RWMutex
	mapping map[string]string
}

func (s *SimpleTracker) Report() map[string]string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	defensiveCopy := make(map[string]string)
	for k, v := range s.mapping {
		defensiveCopy[k] = v
	}
	return defensiveCopy
}

func (s *SimpleTracker) AddReplacement(original string, replacement string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if val, ok := s.mapping[original]; ok {
		if replacement != val {
			klog.Exitf("'%s' already has a value reported as '%s', tried to report '%s'", original, val, replacement)
		}
		return
	}
	s.mapping[original] = replacement
}

func (s *SimpleTracker) GenerateIfAbsent(original string, key string, generator GenerateReplacement) string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if val, ok := s.mapping[original]; ok {
		return val
	}
	if generator == nil {
		return ""
	}
	r := generator(key)
	s.mapping[original] = r
	return r
}

func (s *SimpleTracker) Initialize(replacements map[string]string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if len(s.mapping) > 0 {
		klog.Exitf("tracker was initialized more than once or after some replacements were already added.")
	}
	for k, v := range replacements {
		s.mapping[k] = v
	}
}

func NewSimpleTracker() ReplacementTracker {
	return &SimpleTracker{mapping: map[string]string{}}
}

type StripedTracker struct {
	initialized bool
	stripes     []*sync.Mutex
	mapping     []map[string]string
}

func (s *StripedTracker) Report() map[string]string {
	// this only prevents this method being called concurrently, but not concurrent modifications to the mappings (besides the bucket that hashes from report)
	lock, _ := s.getLockAndMap("report")
	lock.Lock()
	defer lock.Unlock()

	defensiveCopy := make(map[string]string)
	for _, maps := range s.mapping {
		for k, v := range maps {
			defensiveCopy[k] = v
		}
	}
	return defensiveCopy
}

func (s *StripedTracker) AddReplacement(original string, replacement string) {
	lock, mapping := s.getLockAndMap(original)
	lock.Lock()
	defer lock.Unlock()

	if val, ok := mapping[original]; ok {
		if replacement != val {
			klog.Exitf("'%s' already has a value reported as '%s', tried to report '%s'", original, val, replacement)
		}
		return
	}
	mapping[original] = replacement
}

func (s *StripedTracker) GenerateIfAbsent(original string, key string, generator GenerateReplacement) string {
	lock, mapping := s.getLockAndMap(original)
	lock.Lock()
	defer lock.Unlock()

	if val, ok := mapping[original]; ok {
		return val
	}
	if generator == nil {
		return ""
	}
	r := generator(key)
	mapping[original] = r
	return r
}

func (s *StripedTracker) Initialize(replacements map[string]string) {
	lock, _ := s.getLockAndMap("init")
	lock.Lock()
	func(lock *sync.Mutex) {
		defer lock.Unlock()

		if s.initialized {
			klog.Exitf("tracker was initialized more than once or after some replacements were already added.")
		}
		s.initialized = true
	}(lock)

	for k, v := range replacements {
		lock, mapping := s.getLockAndMap(k)
		lock.Lock()
		func(lock *sync.Mutex) {
			defer lock.Unlock()
			mapping[k] = v
		}(lock)
	}
}

// getLockAndMap returns the same stripe (mutex + map entry) for the given key based on its hash.
// We're using a fast 32bit fnv hash to not slow things down too much compared to the savings we get from striping.
func (s *StripedTracker) getLockAndMap(key string) (*sync.Mutex, map[string]string) {
	h := fnv.New32()
	_, _ = h.Write([]byte(key))
	stripeIdx := int(h.Sum32()) % (len(s.stripes))
	return s.stripes[stripeIdx], s.mapping[stripeIdx]
}

func NewStripedTracker() ReplacementTracker {
	// 4*cores performed the best in the setup of the biggest must-gather we could find at 10gb, with about 30k IP addresses
	numStripes := 4 * runtime.NumCPU()
	stripes := make([]*sync.Mutex, numStripes)
	maps := make([]map[string]string, numStripes)
	for i := 0; i < numStripes; i++ {
		stripes[i] = &sync.Mutex{}
		maps[i] = map[string]string{}
	}
	return &StripedTracker{
		stripes: stripes,
		mapping: maps,
	}
}
