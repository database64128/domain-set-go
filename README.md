# domain-set-go

Domain matching in Go.

Most of the code has been upstreamed to [shadowsocks-go](https://github.com/database64128/shadowsocks-go).

## Run Benchmarks

Place your domain set file in the repo directory as `test-domainset.txt`, then run:

```bash
go test -v -benchmem -bench . ./...
```

## License

[AGPLv3](LICENSE)
