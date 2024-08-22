package domainset

import (
	"github.com/database64128/shadowsocks-go/domainset"
)

func BuilderFromTextR(text string) (domainset.Builder, error) {
	return domainset.BuilderFromTextFunc(
		text,
		domainset.NewDomainMapMatcher,
		NewDomainSuffixTrieRMatcherBuilder,
		domainset.NewKeywordLinearMatcher,
		domainset.NewRegexpMatcherBuilder,
	)
}

func BuilderFromTextL(text string) (domainset.Builder, error) {
	return domainset.BuilderFromTextFunc(
		text,
		domainset.NewDomainLinearMatcher,
		domainset.NewSuffixLinearMatcher,
		domainset.NewKeywordLinearMatcher,
		domainset.NewRegexpMatcherBuilder,
	)
}

func DomainSetFromBuilderPreserveType(dsb domainset.Builder) (ds domainset.DomainSet, err error) {
	for _, mb := range dsb {
		if n := mb.MatcherCount(); n == 0 {
			continue
		}
		if m, ok := mb.(domainset.Matcher); ok {
			ds = append(ds, m)
			continue
		}
		ds, err = mb.AppendTo(ds)
		if err != nil {
			return nil, err
		}
	}
	return ds, nil
}
