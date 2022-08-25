package domainset

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/exp/slices"
)

// DomainLinearSet uses a string slice to store suffixes
// and matches domains using linear search.
type DomainLinearSet struct {
	Domains  []string
	Suffixes []string
	Keywords []string
	Regexps  []*regexp.Regexp
}

func NewDomainLinearSet(b []byte) (*DomainLinearSet, error) {
	var ds DomainLinearSet

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
			ds.Domains = append(ds.Domains, line[7:])
		case strings.HasPrefix(line, "suffix:"):
			ds.Suffixes = append(ds.Suffixes, line[7:])
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
func (ds *DomainLinearSet) Match(domain string) bool {
	if slices.Contains(ds.Domains, domain) {
		return true
	}

	for _, suffix := range ds.Suffixes {
		if matchDomainSuffix(domain, suffix) {
			return true
		}
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

func matchDomainSuffix(domain, suffix string) bool {
	return domain == suffix || len(domain) > len(suffix) && domain[len(domain)-len(suffix)-1] == '.' && domain[len(domain)-len(suffix):] == suffix
}
