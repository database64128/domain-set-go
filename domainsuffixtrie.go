package domainset

import (
	"github.com/database64128/shadowsocks-go/domainset"
)

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
	if dt.Children == nil {
		return false
	}

	for i := len(domain) - 1; i >= 0; i-- {
		if domain[i] != '.' {
			continue
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

func (dstr *DomainSuffixTrieR) Rules() []string {
	return (*domainset.DomainSuffixTrie)(dstr).Rules()
}

func (dstr *DomainSuffixTrieR) MatcherCount() int {
	return (*domainset.DomainSuffixTrie)(dstr).MatcherCount()
}

func (dstr *DomainSuffixTrieR) AppendTo(matchers []domainset.Matcher) ([]domainset.Matcher, error) {
	return (*domainset.DomainSuffixTrie)(dstr).AppendTo(matchers)
}

func NewDomainSuffixTrieR(capacity int) domainset.MatcherBuilder {
	return &DomainSuffixTrieR{}
}

func DomainSuffixTrieRFromSlice(suffixes []string) *DomainSuffixTrieR {
	var dstr DomainSuffixTrieR
	for _, s := range suffixes {
		dstr.Insert(s)
	}
	return &dstr
}
