package domainset

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// DomainSuffixMap uses a single map to store all suffixes.
type DomainSuffixMap struct {
	Domains  map[string]struct{}
	Suffixes map[string]struct{}
	Keywords []string
	Regexps  []*regexp.Regexp
}

func NewDomainSuffixMap(b []byte) (*DomainSuffixMap, error) {
	ds := DomainSuffixMap{
		Domains:  make(map[string]struct{}),
		Suffixes: make(map[string]struct{}),
	}

	for {
		lfIndex := bytes.IndexByte(b, '\n')
		if lfIndex == -1 {
			break
		}
		if lfIndex == 0 || lfIndex == 1 && b[0] == '\r' || b[0] == '#' {
			b = b[lfIndex+1:]
			continue
		}
		if b[lfIndex-1] == '\r' {
			lfIndex--
		}
		line := string(b[:lfIndex])
		b = b[lfIndex+1:]

		switch {
		case strings.HasPrefix(line, "domain:"):
			ds.Domains[line[7:]] = struct{}{}
		case strings.HasPrefix(line, "suffix:"):
			ds.Suffixes[line[7:]] = struct{}{}
		case strings.HasPrefix(line, "keyword:"):
			ds.Keywords = append(ds.Keywords, line[8:])
		case strings.HasPrefix(line, "regexp:"):
			regexp, err := regexp.Compile(line[7:])
			if err != nil {
				return nil, err
			}
			ds.Regexps = append(ds.Regexps, regexp)
		default:
			return nil, fmt.Errorf("invalid line: %s", line)
		}
	}

	return &ds, nil
}

// Match returns whether the domain set contains the domain.
func (ds *DomainSuffixMap) Match(domain string) bool {
	if _, ok := ds.Domains[domain]; ok {
		return true
	}

	if ds.matchDomainSuffix(domain) {
		return true
	}

	for _, keyword := range ds.Keywords {
		if strings.Contains(domain, keyword) {
			return true
		}
	}

	for _, regexp := range ds.Regexps {
		if regexp.MatchString(domain) {
			return true
		}
	}

	return false
}

func (ds *DomainSuffixMap) matchDomainSuffix(domain string) bool {
	for i := len(domain) - 1; i >= 0; i-- {
		if domain[i] != '.' {
			continue
		}
		if _, ok := ds.Suffixes[domain[i+1:]]; ok {
			return true
		}
	}
	_, ok := ds.Suffixes[domain]
	return ok
}
