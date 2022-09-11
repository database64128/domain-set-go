package domainset

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/database64128/shadowsocks-go/domainset"
)

type DomainSuffixTrie interface {
	Insert(domain string)
	Match(domain string) bool
}

func DomainSetSuffixTrieFromText(r io.Reader, dst DomainSuffixTrie) (*domainset.DomainSet, error) {
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

	ds := domainset.DomainSet{
		Domains:  make(map[string]struct{}, dskr[0]),
		Suffixes: dst,
		Keywords: make([]string, 0, dskr[2]),
		Regexps:  make([]*regexp.Regexp, 0, dskr[3]),
	}

	for {
		switch {
		case line == "" || strings.IndexByte(line, '#') == 0:
		case strings.HasPrefix(line, domainPrefix):
			ds.Domains[line[domainPrefixLen:]] = struct{}{}
		case strings.HasPrefix(line, suffixPrefix):
			dst.Insert(line[suffixPrefixLen:])
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

func InsertR(dt *domainset.DomainSuffixTrie, domain string) {
	for i := len(domain) - 1; i >= 0; i-- {
		if domain[i] != '.' {
			continue
		}

		part := domain[i+1:]
		var ndt *domainset.DomainSuffixTrie
		if dt.Children == nil {
			ndt = &domainset.DomainSuffixTrie{}
			dt.Children = map[string]*domainset.DomainSuffixTrie{
				part: ndt,
			}
		} else {
			var ok bool
			ndt, ok = dt.Children[part]
			switch {
			case !ok:
				ndt = &domainset.DomainSuffixTrie{}
				dt.Children[part] = ndt
			case ndt.Included:
				return
			}
		}
		InsertR(ndt, domain[:i])
		return
	}

	if dt.Children == nil {
		dt.Children = map[string]*domainset.DomainSuffixTrie{
			domain: {
				Included: true,
			},
		}
	} else {
		ndt, ok := dt.Children[domain]
		if !ok {
			dt.Children[domain] = &domainset.DomainSuffixTrie{
				Included: true,
			}
		} else {
			ndt.Included = true
			ndt.Children = nil
		}
	}
}

func MatchR(dt *domainset.DomainSuffixTrie, domain string) bool {
	for i := len(domain) - 1; i >= 0; i-- {
		if domain[i] != '.' {
			continue
		}

		if dt.Children == nil {
			return false
		}

		ndt, ok := dt.Children[domain[i+1:]]
		if !ok {
			return false
		}
		if ndt.Included {
			return true
		}
		return MatchR(ndt, domain[:i])
	}

	ndt, ok := dt.Children[domain]
	if !ok {
		return false
	}
	return ndt.Included
}

// DomainSuffixTrieR is the same as DomainSuffixTrie,
// but uses recursive algorithms for insertion and search.
type DomainSuffixTrieR domainset.DomainSuffixTrie

func (dstr *DomainSuffixTrieR) Insert(domain string) {
	InsertR((*domainset.DomainSuffixTrie)(dstr), domain)
}

func (dstr *DomainSuffixTrieR) Match(domain string) bool {
	return MatchR((*domainset.DomainSuffixTrie)(dstr), domain)
}
