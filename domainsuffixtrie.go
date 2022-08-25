package domainset

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// DomainTrie is a trie of domain parts segmented by '.'.
type DomainTrie struct {
	included bool
	children map[string]*DomainTrie
}

// Insert inserts a domain suffix to the trie.
// Insertion purges the leaf node's children.
// If say, we insert "www.google.com" and then "google.com",
// The children of node "google" will be purged.
func (dt *DomainTrie) InsertR(domain string) {
	for i := len(domain) - 1; i >= 0; i-- {
		if domain[i] != '.' {
			continue
		}

		part := domain[i+1:]
		var ndt *DomainTrie
		if dt.children == nil {
			ndt = &DomainTrie{}
			dt.children = map[string]*DomainTrie{
				part: ndt,
			}
		} else {
			var ok bool
			ndt, ok = dt.children[part]
			switch {
			case !ok:
				ndt = &DomainTrie{}
				dt.children[part] = ndt
			case ndt.included:
				return
			}
		}
		ndt.InsertR(domain[:i])
		return
	}

	if dt.children == nil {
		dt.children = map[string]*DomainTrie{
			domain: {
				included: true,
			},
		}
	} else {
		ndt, ok := dt.children[domain]
		if !ok {
			dt.children[domain] = &DomainTrie{
				included: true,
			}
		} else {
			ndt.included = true
			ndt.children = nil
		}
	}
}

func (dt *DomainTrie) InsertI(domain string) {
	cdt := dt

	for i := len(domain) - 1; i >= 0; i-- {
		if domain[i] != '.' {
			continue
		}

		part := domain[i+1:]
		domain = domain[:i]
		if cdt.children == nil {
			var ndt DomainTrie
			cdt.children = map[string]*DomainTrie{
				part: &ndt,
			}
			cdt = &ndt
		} else {
			ndt, ok := cdt.children[part]
			switch {
			case !ok:
				ndt = &DomainTrie{}
				cdt.children[part] = ndt
				cdt = ndt
			case ndt.included:
				return
			default:
				cdt = ndt
			}
		}
	}

	if cdt.children == nil {
		cdt.children = map[string]*DomainTrie{
			domain: {
				included: true,
			},
		}
	} else {
		ndt, ok := cdt.children[domain]
		if !ok {
			cdt.children[domain] = &DomainTrie{
				included: true,
			}
		} else {
			ndt.included = true
			ndt.children = nil
		}
	}
}

func (dt *DomainTrie) MatchR(domain string) bool {
	for i := len(domain) - 1; i >= 0; i-- {
		if domain[i] != '.' {
			continue
		}

		if dt.children == nil {
			return false
		}

		ndt, ok := dt.children[domain[i+1:]]
		if !ok {
			return false
		}
		if ndt.included {
			return true
		}
		return ndt.MatchR(domain[:i])
	}

	ndt, ok := dt.children[domain]
	if !ok {
		return false
	}
	return ndt.included
}

func (dt *DomainTrie) MatchI(domain string) bool {
	cdt := dt

	for i := len(domain) - 1; i >= 0; i-- {
		if domain[i] != '.' {
			continue
		}

		if cdt.children == nil {
			return false
		}

		ndt, ok := cdt.children[domain[i+1:]]
		if !ok {
			return false
		}
		if ndt.included {
			return true
		}
		cdt = ndt
		domain = domain[:i]
	}

	ndt, ok := cdt.children[domain]
	if !ok {
		return false
	}
	return ndt.included
}

// DomainSuffixTrie uses a simple trie to store suffixes.
type DomainSuffixTrie struct {
	Domains  map[string]struct{}
	Suffixes *DomainTrie
	Keywords []string
	Regexps  []*regexp.Regexp
}

func NewDomainSuffixTrie(b []byte) (*DomainSuffixTrie, error) {
	ds := DomainSuffixTrie{
		Domains:  make(map[string]struct{}),
		Suffixes: &DomainTrie{},
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
			ds.Suffixes.InsertI(line[7:])
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
func (ds *DomainSuffixTrie) Match(domain string) bool {
	if _, ok := ds.Domains[domain]; ok {
		return true
	}

	if ds.Suffixes.MatchI(domain) {
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

func (ds *DomainSuffixTrie) MatchR(domain string) bool {
	if _, ok := ds.Domains[domain]; ok {
		return true
	}

	if ds.Suffixes.MatchR(domain) {
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
