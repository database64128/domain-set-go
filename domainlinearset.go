package domainset

import (
	"bufio"
	"fmt"
	"io"
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

func DomainLinearSetFromText(r io.Reader) (*DomainLinearSet, error) {
	s := bufio.NewScanner(r)
	if !s.Scan() {
		return nil, errEmptyStream
	}
	line := s.Text()

	dskr, found, err := parseCapacityHint(line)
	if err != nil {
		return nil, err
	}
	if found {
		if !s.Scan() {
			return nil, errEmptyStream
		}
		line = s.Text()
	}

	ds := DomainLinearSet{
		Domains:  make([]string, 0, dskr[0]),
		Suffixes: make([]string, 0, dskr[1]),
		Keywords: make([]string, 0, dskr[2]),
		Regexps:  make([]*regexp.Regexp, 0, dskr[3]),
	}

	for {
		switch {
		case line == "" || strings.IndexByte(line, '#') == 0:
		case strings.HasPrefix(line, domainPrefix):
			ds.Domains = append(ds.Domains, line[domainPrefixLen:])
		case strings.HasPrefix(line, suffixPrefix):
			ds.Suffixes = append(ds.Suffixes, line[suffixPrefixLen:])
		case strings.HasPrefix(line, keywordPrefix):
			ds.Keywords = append(ds.Keywords, line[keywordPrefixLen:])
		case strings.HasPrefix(line, regexpPrefix):
			regexp, err := regexp.Compile(line[regexpPrefixLen:])
			if err != nil {
				return nil, err
			}
			ds.Regexps = append(ds.Regexps, regexp)
		default:
			return nil, fmt.Errorf("invalid line: %s", line)
		}

		if !s.Scan() {
			break
		}
		line = s.Text()
	}

	return &ds, nil
}
