package domainset

import (
	"os"
	"testing"
)

const (
	filename     = "test-domainset.txt"
	shortDomain  = "localhost"
	mediumDomain = "www.example.com"
	longDomain   = "cant.come.up.with.a.long.domain.name"
	testDS       = `
# Test comment.
domain:www.example.net
suffix:example.com
suffix:youtube.com
suffix:gen.xyz
suffix:cube64128.xyz
suffix:dynmap.us
suffix:us
suffix:about.us
keyword:org
regexp:^adservice\.google\.([a-z]{2}|com?)(\.[a-z]{2})?$
`
)

var (
	data        []byte
	testDSBytes = []byte(testDS)
)

func init() {
	var err error
	data, err = os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
}

func testMatch(t *testing.T, ds DomainSet, domain string, expectedResult bool) {
	if ds.Match(domain) != expectedResult {
		t.Errorf("%s should return %v", domain, expectedResult)
	}
}

func testDomainSet(t *testing.T, ds DomainSet) {
	testMatch(t, ds, "example.net", false)
	testMatch(t, ds, "www.example.net", true)
	testMatch(t, ds, "example.com", true)
	testMatch(t, ds, "www.example.com", true)
	testMatch(t, ds, "gobyexample.com", false)
	testMatch(t, ds, "example.org", true)
	testMatch(t, ds, "adservice.google.com", true)
}

func TestDomainLinearSet(t *testing.T) {
	ds, err := NewDomainLinearSet(testDSBytes)
	if err != nil {
		t.Fatal(err)
	}
	testDomainSet(t, ds)
}

func TestDomainSuffixMap(t *testing.T) {
	ds, err := NewDomainSuffixMap(testDSBytes)
	if err != nil {
		t.Fatal(err)
	}
	testDomainSet(t, ds)
}

func TestDomainSuffixTrie(t *testing.T) {
	ds, err := NewDomainSuffixTrie(testDSBytes)
	if err != nil {
		t.Fatal(err)
	}
	testDomainSet(t, ds)
}

func BenchmarkDomainLinearSetSetup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := NewDomainLinearSet(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDomainLinearSetMatch(b *testing.B) {
	ds, err := NewDomainLinearSet(data)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("short", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ds.Match(shortDomain)
		}
	})

	b.Run("medium", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ds.Match(mediumDomain)
		}
	})

	b.Run("long", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ds.Match(longDomain)
		}
	})
}

func BenchmarkDomainSuffixMapSetup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := NewDomainSuffixMap(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDomainSuffixMapMatch(b *testing.B) {
	ds, err := NewDomainSuffixMap(data)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("short", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ds.Match(shortDomain)
		}
	})

	b.Run("medium", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ds.Match(mediumDomain)
		}
	})

	b.Run("long", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ds.Match(longDomain)
		}
	})
}

func BenchmarkDomainSuffixTrieSetup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := NewDomainSuffixTrie(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDomainSuffixTrieMatchRecursion(b *testing.B) {
	ds, err := NewDomainSuffixTrie(data)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("short", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ds.MatchR(shortDomain)
		}
	})

	b.Run("medium", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ds.MatchR(mediumDomain)
		}
	})

	b.Run("long", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ds.MatchR(longDomain)
		}
	})
}

func BenchmarkDomainSuffixTrieMatchIteration(b *testing.B) {
	ds, err := NewDomainSuffixTrie(data)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("short", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ds.Match(shortDomain)
		}
	})

	b.Run("medium", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ds.Match(mediumDomain)
		}
	})

	b.Run("long", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ds.Match(longDomain)
		}
	})
}
