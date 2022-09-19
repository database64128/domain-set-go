package domainset

import (
	"io"

	"github.com/database64128/shadowsocks-go/domainset"
)

func BuilderFromTextR(r io.Reader) (domainset.Builder, error) {
	return domainset.BuilderFromTextFunc(
		r,
		domainset.NewDomainMapMatcher,
		NewDomainSuffixTrieR,
		domainset.NewKeywordLinearMatcher,
		domainset.NewRegexpMatcherBuilder,
	)
}

func BuilderFromTextL(r io.Reader) (domainset.Builder, error) {
	return domainset.BuilderFromTextFunc(
		r,
		domainset.NewDomainLinearMatcher,
		domainset.NewSuffixLinearMatcher,
		domainset.NewKeywordLinearMatcher,
		domainset.NewRegexpMatcherBuilder,
	)
}
