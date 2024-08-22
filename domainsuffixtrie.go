package domainset

import (
	"strings"

	"github.com/database64128/shadowsocks-go/domainset"
)

func InsertR(dt domainset.DomainSuffixTrie, domain string) {
	dotIndex := strings.LastIndexByte(domain, '.')
	if dotIndex == -1 {
		// Make the final (from right to left) part a leaf node.
		dt.Children[domain] = domainset.DomainSuffixTrie{}
		return
	}

	part := domain[dotIndex+1:]

	ndt, ok := dt.Children[part]
	switch {
	case !ok:
		ndt = domainset.DomainSuffixTrie{
			Children: make(map[string]domainset.DomainSuffixTrie, 1),
		}
		dt.Children[part] = ndt
	case ndt.Children == nil:
		return
	}

	InsertR(ndt, domain[:dotIndex])
}

func MatchR(dt domainset.DomainSuffixTrie, domain string) bool {
	dotIndex := strings.LastIndexByte(domain, '.')
	part := domain[dotIndex+1:] // If dotIndex == -1, part will be domain.
	ndt, ok := dt.Children[part]
	if !ok {
		return false
	}
	if ndt.Children == nil {
		return true
	}
	if dotIndex == -1 {
		return false
	}
	return MatchR(ndt, domain[:dotIndex])
}

// DomainSuffixTrieR is like [domainset.DomainSuffixTrie],
// but uses recursive algorithms for insertion and search.
type DomainSuffixTrieR struct {
	domainset.DomainSuffixTrie
}

func (dstr DomainSuffixTrieR) Insert(domain string) {
	InsertR(dstr.DomainSuffixTrie, domain)
}

func (dstr DomainSuffixTrieR) Match(domain string) bool {
	return MatchR(dstr.DomainSuffixTrie, domain)
}

func (dstr *DomainSuffixTrieR) AppendTo(matchers []domainset.Matcher) ([]domainset.Matcher, error) {
	if len(dstr.DomainSuffixTrie.Children) == 0 {
		return matchers, nil
	}
	return append(matchers, dstr), nil
}

func NewDomainSuffixTrieRMatcherBuilder(_ int) domainset.MatcherBuilder {
	return &DomainSuffixTrieR{
		DomainSuffixTrie: domainset.NewDomainSuffixTrie(),
	}
}
