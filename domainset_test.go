package domainset

import (
	"bytes"
	"os"
	"testing"
	"unsafe"

	"github.com/database64128/shadowsocks-go/domainset"
	"github.com/database64128/shadowsocks-go/mmap"
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

var data string

func TestMain(m *testing.M) {
	var (
		close func() error
		err   error
	)
	data, close, err = mmap.ReadFile[string](filename)
	if err != nil {
		panic(err)
	}
	defer close()
	m.Run()
}

func testMatch(t *testing.T, ds domainset.DomainSet, domain string, expectedResult bool) {
	t.Helper()
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
	dsb, err := BuilderFromTextR(testDS)
	if err != nil {
		t.Fatal(err)
	}
	ds, err := DomainSetFromBuilderPreserveType(dsb)
	if err != nil {
		t.Fatal(err)
	}
	testDomainSet(t, ds)
}

func benchmarkDomainSetSetup(b *testing.B, setup func(string) (domainset.Builder, error)) {
	for range b.N {
		if _, err := setup(data); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkDomainSetMatch(b *testing.B, setup func(string) (domainset.Builder, error)) {
	dsb, err := setup(data)
	if err != nil {
		b.Fatal(err)
	}
	ds, err := DomainSetFromBuilderPreserveType(dsb)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("Short", func(b *testing.B) {
		for range b.N {
			_ = ds.Match(shortDomain)
		}
	})

	b.Run("Medium", func(b *testing.B) {
		for range b.N {
			_ = ds.Match(mediumDomain)
		}
	})

	b.Run("Long", func(b *testing.B) {
		for range b.N {
			_ = ds.Match(longDomain)
		}
	})
}

func BenchmarkDomainSetSetup(b *testing.B) {
	for _, c := range []struct {
		name  string
		setup func(string) (domainset.Builder, error)
	}{
		{"Linear", BuilderFromTextLinear},
		{"LinearClone", BuilderFromTextLinearClone},
		{"LinearBulkClone", BuilderFromTextLinearBulkClone},
		{"LinearIter", BuilderFromTextLinearIter},
		{"SuffixMap", domainset.BuilderFromTextFast},
		{"SuffixMapClone", domainset.BuilderFromTextFastClone},
		{"SuffixMapBulkClone", BuilderFromTextFastBulkClone},
		{"SuffixTrieIteration", domainset.BuilderFromText},
		{"SuffixTrieIterationClone", domainset.BuilderFromTextClone},
		{"SuffixTrieIterationBulkClone", BuilderFromTextBulkClone},
		{"SuffixTrieRecursion", BuilderFromTextR},
	} {
		b.Run(c.name, func(b *testing.B) {
			benchmarkDomainSetSetup(b, c.setup)
		})
	}
}

func BenchmarkDomainSetMatch(b *testing.B) {
	for _, c := range []struct {
		name  string
		setup func(string) (domainset.Builder, error)
	}{
		{"Linear", BuilderFromTextLinear},
		{"SuffixMap", domainset.BuilderFromTextFast},
		{"SuffixTrieIteration", domainset.BuilderFromText},
		{"SuffixTrieRecursion", BuilderFromTextR},
	} {
		b.Run(c.name, func(b *testing.B) {
			benchmarkDomainSetMatch(b, c.setup)
		})
	}
}

func BenchmarkDomainSuffixTrieSetupIterationTextMmapBulkClone(b *testing.B) {
	for range b.N {
		data, close, err := mmap.ReadFile[string](filename)
		if err != nil {
			b.Fatal(err)
		}
		defer close()

		if _, err := BuilderFromTextBulkClone(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDomainSuffixTrieSetupIterationTextReadAll(b *testing.B) {
	for range b.N {
		data, err := os.ReadFile(filename)
		if err != nil {
			b.Fatal(err)
		}

		if _, err := domainset.BuilderFromText(unsafe.String(unsafe.SliceData(data), len(data))); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDomainSuffixTrieSetupIterationGob(b *testing.B) {
	dsb, err := domainset.BuilderFromText(data)
	if err != nil {
		b.Fatal(err)
	}

	var buffer bytes.Buffer
	if err := dsb.WriteGob(&buffer); err != nil {
		b.Fatal(err)
	}
	buf := buffer.Bytes()
	r := bytes.NewReader(buf)
	b.Logf("gob encoded size: %d", len(buf))
	b.ResetTimer()

	for range b.N {
		if _, err := domainset.BuilderFromGob(r); err != nil {
			b.Fatal(err)
		}
		r.Reset(buf)
	}
}
