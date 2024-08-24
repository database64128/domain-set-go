package domainset

import (
	"errors"
	"fmt"
	"strings"

	"github.com/database64128/shadowsocks-go/bytestrings"
	"github.com/database64128/shadowsocks-go/domainset"
)

const (
	domainPrefix     = "domain:"
	suffixPrefix     = "suffix:"
	keywordPrefix    = "keyword:"
	regexpPrefix     = "regexp:"
	domainPrefixLen  = len(domainPrefix)
	suffixPrefixLen  = len(suffixPrefix)
	keywordPrefixLen = len(keywordPrefix)
	regexpPrefixLen  = len(regexpPrefix)
)

var errEmptySet = errors.New("empty domain set")

func noCloneString(s string) string {
	return s
}

// bulkStringCloner uses a large contiguous buffer to efficiently clone many small strings.
type bulkStringCloner struct {
	sb strings.Builder
}

// newBulkStringCloner creates a [bulkStringCloner] with the given capacity.
func newBulkStringCloner(capacity int) *bulkStringCloner {
	c := &bulkStringCloner{}
	c.sb.Grow(capacity)
	return c
}

// reserve reserves buffer space for n bytes.
func (c *bulkStringCloner) reserve(n int) {
	if bufCap := c.sb.Cap(); bufCap-c.sb.Len() < n {
		c.sb.Reset()
		c.sb.Grow(max(bufCap, n))
	}
}

// Clone clones the given string.
func (c *bulkStringCloner) Clone(s string) string {
	c.reserve(len(s))
	bufLen := c.sb.Len()
	_, _ = c.sb.WriteString(s)
	return c.sb.String()[bufLen:]
}

func BuilderFromTextLinear(text string) (domainset.Builder, error) {
	return domainset.BuilderFromTextFunc(
		text,
		domainset.NewDomainLinearMatcher,
		domainset.NewSuffixLinearMatcher,
		domainset.NewKeywordLinearMatcher,
		domainset.NewRegexpMatcherBuilder,
		noCloneString,
	)
}

func BuilderFromTextLinearClone(text string) (domainset.Builder, error) {
	return domainset.BuilderFromTextFunc(
		text,
		domainset.NewDomainLinearMatcher,
		domainset.NewSuffixLinearMatcher,
		domainset.NewKeywordLinearMatcher,
		domainset.NewRegexpMatcherBuilder,
		strings.Clone,
	)
}

func BuilderFromTextLinearBulkClone(text string) (domainset.Builder, error) {
	c := newBulkStringCloner(len(text) / 4)
	return domainset.BuilderFromTextFunc(
		text,
		domainset.NewDomainLinearMatcher,
		domainset.NewSuffixLinearMatcher,
		domainset.NewKeywordLinearMatcher,
		domainset.NewRegexpMatcherBuilder,
		c.Clone,
	)
}

func BuilderFromTextLinearIter(text string) (domainset.Builder, error) {
	return BuilderFromTextFunc(
		text,
		domainset.NewDomainLinearMatcher,
		domainset.NewSuffixLinearMatcher,
		domainset.NewKeywordLinearMatcher,
		domainset.NewRegexpMatcherBuilder,
	)
}

func BuilderFromTextFastBulkClone(text string) (domainset.Builder, error) {
	c := newBulkStringCloner(len(text) / 4)
	return domainset.BuilderFromTextFunc(
		text,
		domainset.NewDomainMapMatcher,
		domainset.NewSuffixMapMatcher,
		domainset.NewKeywordLinearMatcher,
		domainset.NewRegexpMatcherBuilder,
		c.Clone,
	)
}

func BuilderFromTextBulkClone(text string) (domainset.Builder, error) {
	c := newBulkStringCloner(len(text) / 4)
	return domainset.BuilderFromTextFunc(
		text,
		domainset.NewDomainMapMatcher,
		domainset.NewDomainSuffixTrieMatcherBuilder,
		domainset.NewKeywordLinearMatcher,
		domainset.NewRegexpMatcherBuilder,
		c.Clone,
	)
}

func BuilderFromTextR(text string) (domainset.Builder, error) {
	return BuilderFromTextFunc(
		text,
		domainset.NewDomainMapMatcher,
		NewDomainSuffixTrieRMatcherBuilder,
		domainset.NewKeywordLinearMatcher,
		domainset.NewRegexpMatcherBuilder,
	)
}

func BuilderFromTextFunc(
	text string,
	newDomainMatcherBuilder,
	newSuffixMatcherBuilder,
	newKeywordMatcherBuilder,
	newRegexpMatcherBuilder func(int) domainset.MatcherBuilder,
) (domainset.Builder, error) {
	line, remaining := bytestrings.NextNonEmptyLine(text)
	if len(line) == 0 {
		return domainset.Builder{}, errEmptySet
	}

	dskr, found, err := domainset.ParseCapacityHint(line)
	if err != nil {
		return domainset.Builder{}, err
	}
	if !found {
		remaining = text
	}

	dsb := domainset.Builder{
		newDomainMatcherBuilder(dskr[0]),
		newSuffixMatcherBuilder(dskr[1]),
		newKeywordMatcherBuilder(dskr[2]),
		newRegexpMatcherBuilder(dskr[3]),
	}

	for line = range bytestrings.NonEmptyLines(remaining) {
		// domainPrefixLen == suffixPrefixLen == regexpPrefixLen == 7
		if len(line) > 7 {
			switch line[:7] {
			case suffixPrefix:
				dsb.SuffixMatcherBuilder().Insert(line[7:])
				continue
			case domainPrefix:
				dsb.DomainMatcherBuilder().Insert(line[7:])
				continue
			case regexpPrefix:
				dsb.RegexpMatcherBuilder().Insert(line[7:])
				continue
			case keywordPrefix[:7]:
				if len(line) <= keywordPrefixLen || line[7] != keywordPrefix[7] {
					return dsb, fmt.Errorf("invalid line: %q", line)
				}
				dsb.KeywordMatcherBuilder().Insert(line[keywordPrefixLen:])
				continue
			}
		}

		if line[0] != '#' {
			return dsb, fmt.Errorf("invalid line: %q", line)
		}
	}

	return dsb, nil
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
