# Testing

Unit tests
- Run: `make test` (short, race enabled via Makefile)
- Fakes used for ports; deterministic audio fixtures
- Coverage gates in CI: total ≥ 85%, usecase ≥ 90%

Integration tests
- Build whisper lib: `make deps`
- Run: `GOSPER_INTEGRATION=1 go test ./test/integration -tags whisper -v`
- Provide model path via `GOSPER_MODEL_PATH`
- CLI e2e smoke also requires `-tags "cli whisper"`

Golden tests
- `test/testdata/golden/` holds texts; tolerant overlap matching by default

Tips
- Set `GOMAXPROCS=1` to expose concurrency bugs
- Use `-run` to limit scope during development

