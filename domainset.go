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
	return BuilderFromTextFunc(
		text,
		domainset.NewDomainLinearMatcher,
		domainset.NewSuffixLinearMatcher,
		domainset.NewKeywordLinearMatcher,
		domainset.NewRegexpMatcherBuilder,
	)
}

func BuilderFromTextLinearClone(text string) (domainset.Builder, error) {
	return BuilderFromTextCloneFunc(
		text,
		domainset.NewDomainLinearMatcher,
		domainset.NewSuffixLinearMatcher,
		domainset.NewKeywordLinearMatcher,
		domainset.NewRegexpMatcherBuilder,
	)
}

func BuilderFromTextLinearBulkClone(text string) (domainset.Builder, error) {
	return BuilderFromTextBulkCloneFunc(
		text,
		domainset.NewDomainLinearMatcher,
		domainset.NewSuffixLinearMatcher,
		domainset.NewKeywordLinearMatcher,
		domainset.NewRegexpMatcherBuilder,
	)
}

func BuilderFromTextLinearIter(text string) (domainset.Builder, error) {
	return BuilderFromTextIterFunc(
		text,
		domainset.NewDomainLinearMatcher,
		domainset.NewSuffixLinearMatcher,
		domainset.NewKeywordLinearMatcher,
		domainset.NewRegexpMatcherBuilder,
	)
}

// BuilderFromTextFast is like [BuilderFromText], but prefers the [domainset.SuffixMapMatcher] for suffix matching.
// It's only faster when building the matcher. The resulting matcher is actually a bit slower.
func BuilderFromTextFast(text string) (domainset.Builder, error) {
	return BuilderFromTextFunc(
		text,
		domainset.NewDomainMapMatcher,
		domainset.NewSuffixMapMatcher,
		domainset.NewKeywordLinearMatcher,
		domainset.NewRegexpMatcherBuilder,
	)
}

// BuilderFromTextFastClone is like [BuilderFromTextFast], but clones the rule strings.
// The returned builder has no reference to the input text.
func BuilderFromTextFastClone(text string) (domainset.Builder, error) {
	return BuilderFromTextCloneFunc(
		text,
		domainset.NewDomainMapMatcher,
		domainset.NewSuffixMapMatcher,
		domainset.NewKeywordLinearMatcher,
		domainset.NewRegexpMatcherBuilder,
	)
}

// BuilderFromTextFastBulkClone is like [BuilderFromTextFastClone], but uses a bulk string cloner.
func BuilderFromTextFastBulkClone(text string) (domainset.Builder, error) {
	return BuilderFromTextBulkCloneFunc(
		text,
		domainset.NewDomainMapMatcher,
		domainset.NewSuffixMapMatcher,
		domainset.NewKeywordLinearMatcher,
		domainset.NewRegexpMatcherBuilder,
	)
}

// BuilderFromTextClone is like [BuilderFromText], but clones the rule strings.
// The returned builder has no reference to the input text.
func BuilderFromTextClone(text string) (domainset.Builder, error) {
	return BuilderFromTextCloneFunc(
		text,
		domainset.NewDomainMapMatcher,
		domainset.NewDomainSuffixTrieMatcherBuilder,
		domainset.NewKeywordLinearMatcher,
		domainset.NewRegexpMatcherBuilder,
	)
}

// BuilderFromTextBulkClone is like [BuilderFromTextClone], but uses a bulk string cloner.
func BuilderFromTextBulkClone(text string) (domainset.Builder, error) {
	return BuilderFromTextBulkCloneFunc(
		text,
		domainset.NewDomainMapMatcher,
		domainset.NewDomainSuffixTrieMatcherBuilder,
		domainset.NewKeywordLinearMatcher,
		domainset.NewRegexpMatcherBuilder,
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

// BuilderFromTextFunc parses the text for domain set rules,
// inserts them into matcher builders created by the given functions,
// and returns the resulting builder.
func BuilderFromTextFunc(
	text string,
	newDomainMatcherBuilder,
	newSuffixMatcherBuilder,
	newKeywordMatcherBuilder,
	newRegexpMatcherBuilder func(int) domainset.MatcherBuilder,
) (domainset.Builder, error) {
	line, text := bytestrings.NextNonEmptyLine(text)
	if len(line) == 0 {
		return domainset.Builder{}, errEmptySet
	}

	dskr, found, err := domainset.ParseCapacityHint(line)
	if err != nil {
		return domainset.Builder{}, err
	}
	if found {
		line, text = bytestrings.NextNonEmptyLine(text)
		if len(line) == 0 {
			return domainset.Builder{}, errEmptySet
		}
	}

	dsb := domainset.Builder{
		newDomainMatcherBuilder(dskr[0]),
		newSuffixMatcherBuilder(dskr[1]),
		newKeywordMatcherBuilder(dskr[2]),
		newRegexpMatcherBuilder(dskr[3]),
	}

	for {
		// domainPrefixLen == suffixPrefixLen == regexpPrefixLen == 7
		if len(line) > 7 {
			switch line[:7] {
			case suffixPrefix:
				dsb.SuffixMatcherBuilder().Insert(line[suffixPrefixLen:])
				goto next
			case domainPrefix:
				dsb.DomainMatcherBuilder().Insert(line[domainPrefixLen:])
				goto next
			case regexpPrefix:
				dsb.RegexpMatcherBuilder().Insert(line[regexpPrefixLen:])
				goto next
			case keywordPrefix[:7]:
				if len(line) <= keywordPrefixLen || line[7] != keywordPrefix[7] {
					return dsb, fmt.Errorf("invalid line: %q", line)
				}
				dsb.KeywordMatcherBuilder().Insert(line[keywordPrefixLen:])
				goto next
			}
		}

		if line[0] != '#' {
			return dsb, fmt.Errorf("invalid line: %q", line)
		}

	next:
		line, text = bytestrings.NextNonEmptyLine(text)
		if len(line) == 0 {
			break
		}
	}

	return dsb, nil
}

// BuilderFromTextCloneFunc is like [BuilderFromTextFunc], but clones the rule strings.
func BuilderFromTextCloneFunc(
	text string,
	newDomainMatcherBuilder,
	newSuffixMatcherBuilder,
	newKeywordMatcherBuilder,
	newRegexpMatcherBuilder func(int) domainset.MatcherBuilder,
) (domainset.Builder, error) {
	line, text := bytestrings.NextNonEmptyLine(text)
	if len(line) == 0 {
		return domainset.Builder{}, errEmptySet
	}

	dskr, found, err := domainset.ParseCapacityHint(line)
	if err != nil {
		return domainset.Builder{}, err
	}
	if found {
		line, text = bytestrings.NextNonEmptyLine(text)
		if len(line) == 0 {
			return domainset.Builder{}, errEmptySet
		}
	}

	dsb := domainset.Builder{
		newDomainMatcherBuilder(dskr[0]),
		newSuffixMatcherBuilder(dskr[1]),
		newKeywordMatcherBuilder(dskr[2]),
		newRegexpMatcherBuilder(dskr[3]),
	}

	for {
		// domainPrefixLen == suffixPrefixLen == regexpPrefixLen == 7
		if len(line) > 7 {
			switch line[:7] {
			case suffixPrefix:
				dsb.SuffixMatcherBuilder().Insert(strings.Clone(line[suffixPrefixLen:]))
				goto next
			case domainPrefix:
				dsb.DomainMatcherBuilder().Insert(strings.Clone(line[domainPrefixLen:]))
				goto next
			case regexpPrefix:
				dsb.RegexpMatcherBuilder().Insert(strings.Clone(line[regexpPrefixLen:]))
				goto next
			case keywordPrefix[:7]:
				if len(line) <= keywordPrefixLen || line[7] != keywordPrefix[7] {
					return dsb, fmt.Errorf("invalid line: %q", line)
				}
				dsb.KeywordMatcherBuilder().Insert(strings.Clone(line[keywordPrefixLen:]))
				goto next
			}
		}

		if line[0] != '#' {
			return dsb, fmt.Errorf("invalid line: %q", line)
		}

	next:
		line, text = bytestrings.NextNonEmptyLine(text)
		if len(line) == 0 {
			break
		}
	}

	return dsb, nil
}

// BuilderFromTextBulkCloneFunc is like [BuilderFromTextCloneFunc], but uses a bulk string cloner.
func BuilderFromTextBulkCloneFunc(
	text string,
	newDomainMatcherBuilder,
	newSuffixMatcherBuilder,
	newKeywordMatcherBuilder,
	newRegexpMatcherBuilder func(int) domainset.MatcherBuilder,
) (domainset.Builder, error) {
	line, text := bytestrings.NextNonEmptyLine(text)
	if len(line) == 0 {
		return domainset.Builder{}, errEmptySet
	}

	dskr, found, err := domainset.ParseCapacityHint(line)
	if err != nil {
		return domainset.Builder{}, err
	}
	if found {
		line, text = bytestrings.NextNonEmptyLine(text)
		if len(line) == 0 {
			return domainset.Builder{}, errEmptySet
		}
	}

	dsb := domainset.Builder{
		newDomainMatcherBuilder(dskr[0]),
		newSuffixMatcherBuilder(dskr[1]),
		newKeywordMatcherBuilder(dskr[2]),
		newRegexpMatcherBuilder(dskr[3]),
	}

	c := newBulkStringCloner(len(text) / 4)

	for {
		// domainPrefixLen == suffixPrefixLen == regexpPrefixLen == 7
		if len(line) > 7 {
			switch line[:7] {
			case suffixPrefix:
				dsb.SuffixMatcherBuilder().Insert(c.Clone(line[suffixPrefixLen:]))
				goto next
			case domainPrefix:
				dsb.DomainMatcherBuilder().Insert(c.Clone(line[domainPrefixLen:]))
				goto next
			case regexpPrefix:
				dsb.RegexpMatcherBuilder().Insert(c.Clone(line[regexpPrefixLen:]))
				goto next
			case keywordPrefix[:7]:
				if len(line) <= keywordPrefixLen || line[7] != keywordPrefix[7] {
					return dsb, fmt.Errorf("invalid line: %q", line)
				}
				dsb.KeywordMatcherBuilder().Insert(c.Clone(line[keywordPrefixLen:]))
				goto next
			}
		}

		if line[0] != '#' {
			return dsb, fmt.Errorf("invalid line: %q", line)
		}

	next:
		line, text = bytestrings.NextNonEmptyLine(text)
		if len(line) == 0 {
			break
		}
	}

	return dsb, nil
}

// BuilderFromTextIterFunc is like [BuilderFromTextFunc], but uses an iterator to separate lines.
func BuilderFromTextIterFunc(
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
				dsb.SuffixMatcherBuilder().Insert(line[suffixPrefixLen:])
				continue
			case domainPrefix:
				dsb.DomainMatcherBuilder().Insert(line[domainPrefixLen:])
				continue
			case regexpPrefix:
				dsb.RegexpMatcherBuilder().Insert(line[regexpPrefixLen:])
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

// DomainSetFromBuilderPreserveType is like [domainset.Builder.DomainSet],
// but avoids matcher conversions when possible. This is useful for testing
// and benchmarking different matcher types.
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
