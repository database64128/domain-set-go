package domainset

// DomainSet is a rule set that can match domains.
type DomainSet interface {
	Match(domain string) bool
}
