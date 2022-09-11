package domainset

import (
	"errors"
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

var errEmptyStream = errors.New("empty stream")

// DomainSet is a rule set that can match domains.
type DomainSet interface {
	Match(domain string) bool
}
