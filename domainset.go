package domainset

import (
	"github.com/database64128/shadowsocks-go/domainset"
)

func BuilderFromTextR(text string) (domainset.Builder, error) {
	return domainset.BuilderFromTextFunc(
		text,
		domainset.NewDomainMapMatcher,
		NewDomainSuffixTrieR,
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
