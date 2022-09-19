package domainset

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/database64128/shadowsocks-go/domainset"
)

const (
	filename     = "test-domainset.txt"
	shortDomain  = "localhost"
	mediumDomain = "www.example.com"
	longDomain   = "cant.come.up.with.a.long.domain.name"
	testDS       = `# shadowsocks-go domain set capacity hint 1 6 1 1 DSKR
domain:www.example.net
suffix:example.com
suffix:github.com
suffix:cube64128.xyz
suffix:api.ipify.org
suffix:api6.ipify.org
suffix:archlinux.org
keyword:dev
regexp:^adservice\.google\.([a-z]{2}|com?)(\.[a-z]{2})?$
`
)

var data []byte

func init() {
	var err error
	data, err = os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
}

func testMatch(t *testing.T, ds domainset.DomainSet, domain string, expectedResult bool) {
	if ds.Match(domain) != expectedResult {
		t.Errorf("%s should return %v", domain, expectedResult)
	}
}

func testDomainSet(t *testing.T, ds domainset.DomainSet) {
	testMatch(t, ds, "net", false)
	testMatch(t, ds, "example.net", false)
	testMatch(t, ds, "www.example.net", true)
	testMatch(t, ds, "wwww.example.net", false)
	testMatch(t, ds, "test.www.example.net", false)
	testMatch(t, ds, "com", false)
	testMatch(t, ds, "example.com", true)
	testMatch(t, ds, "www.example.com", true)
	testMatch(t, ds, "gobyexample.com", false)
	testMatch(t, ds, "example.org", false)
	testMatch(t, ds, "github.com", true)
	testMatch(t, ds, "api.github.com", true)
	testMatch(t, ds, "raw.githubusercontent.com", false)
	testMatch(t, ds, "github.blog", false)
	testMatch(t, ds, "cube64128.xyz", true)
	testMatch(t, ds, "www.cube64128.xyz", true)
	testMatch(t, ds, "notcube64128.xyz", false)
	testMatch(t, ds, "org", false)
	testMatch(t, ds, "ipify.org", false)
	testMatch(t, ds, "api.ipify.org", true)
	testMatch(t, ds, "api6.ipify.org", true)
	testMatch(t, ds, "api64.ipify.org", false)
	testMatch(t, ds, "www.ipify.org", false)
	testMatch(t, ds, "archlinux.org", true)
	testMatch(t, ds, "aur.archlinux.org", true)
	testMatch(t, ds, "wikipedia.org", false)
	testMatch(t, ds, "dev", true)
	testMatch(t, ds, "go.dev", true)
	testMatch(t, ds, "drewdevault.com", true)
	testMatch(t, ds, "developer.mozilla.org", true)
	testMatch(t, ds, "adservice.google.com", true)
}

func TestDomainSuffixTrieR(t *testing.T) {
	r := strings.NewReader(testDS)
	dsb, err := BuilderFromTextR(r)
	if err != nil {
		t.Fatal(err)
	}
	ds, err := dsb.DomainSet()
	if err != nil {
		t.Fatal(err)
	}
	testDomainSet(t, ds)
}

func benchmarkDomainSetSetup(b *testing.B, setup func(r io.Reader) (domainset.Builder, error)) {
	r := bytes.NewReader(data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := setup(r); err != nil {
			b.Fatal(err)
		}
		r.Reset(data)
	}
}

func benchmarkDomainSetMatch(b *testing.B, setup func(r io.Reader) (domainset.Builder, error)) {
	r := bytes.NewReader(data)
	dsb, err := setup(r)
	if err != nil {
		b.Fatal(err)
	}
	ds, err := dsb.DomainSet()
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

func BenchmarkDomainLinearSetSetup(b *testing.B) {
	benchmarkDomainSetSetup(b, BuilderFromTextL)
}

func BenchmarkDomainLinearSetMatch(b *testing.B) {
	benchmarkDomainSetMatch(b, BuilderFromTextL)
}

func BenchmarkDomainSuffixMapSetup(b *testing.B) {
	benchmarkDomainSetSetup(b, domainset.BuilderFromTextFast)
}

func BenchmarkDomainSuffixMapMatch(b *testing.B) {
	benchmarkDomainSetMatch(b, domainset.BuilderFromTextFast)
}

func BenchmarkDomainSuffixTrieSetupIteration(b *testing.B) {
	benchmarkDomainSetSetup(b, domainset.BuilderFromText)
}

func BenchmarkDomainSuffixTrieSetupRecursion(b *testing.B) {
	benchmarkDomainSetSetup(b, BuilderFromTextR)
}

func BenchmarkDomainSuffixTrieMatchIteration(b *testing.B) {
	benchmarkDomainSetMatch(b, domainset.BuilderFromText)
}

func BenchmarkDomainSuffixTrieMatchRecursion(b *testing.B) {
	benchmarkDomainSetMatch(b, BuilderFromTextR)
}

func BenchmarkDomainSuffixTrieSetupIterationGob(b *testing.B) {
	r := bytes.NewReader(data)
	dsb, err := domainset.BuilderFromText(r)
	if err != nil {
		b.Fatal(err)
	}

	var buffer bytes.Buffer
	if err := dsb.WriteGob(&buffer); err != nil {
		b.Fatal(err)
	}
	buf := buffer.Bytes()
	r.Reset(buf)
	b.Logf("gob encoded size: %d", len(buf))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := domainset.BuilderFromGob(r); err != nil {
			b.Fatal(err)
		}
		r.Reset(buf)
	}
}
