# CI

GitHub Actions
- Matrix: macos-latest, ubuntu-latest
- Steps: build libwhisper (bindings/go), run unit tests, coverage gates

Coverage gates
- Total coverage ≥ 85%
- Usecase package coverage ≥ 90%

Integration tests
- Kept out of default CI runs; opt-in via `GOSPER_INTEGRATION=1` and tags

