Contributing to imUp
--

[Fork the repository](https://github.com/imup-io/client/fork) on and clone the repository onto your machine.

Pull requests are welcome! For major changes, please open an issue first to ensure your change aligns with our overall road-map.

Please update tests and include code comments where appropriate.

## Running Tests

Note that this project is currently not running tests in CI, and the default
test suite assumes current connectivity to the internet.

```go
go test -race -coverage=c.out ./...
```

## Coverage

There is no "quality" gate enforced in CI around coverage due to some limitations
with running the existing testing suite in github actions, however this is the projects
current (as of this commit) coverage.

```sh
ok   github.com/imup-io/client 69.898s coverage: 71.9% of statements
ok   github.com/imup-io/client/config 0.289s coverage: 74.5% of statements
ok   github.com/imup-io/client/util 0.458s coverage: 92.3% of statements
```
