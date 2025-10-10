# Contributing

Workflow
- Create a feature branch: `NNN-short-name`
- Keep commits focused and descriptive (spec/plan/tasks → scaffolding → feature → tests → deploy)
- Run `make test` locally before PRs

Style
- Go 1.22+, idiomatic Go, table-driven tests
- Keep use cases free of side effects (call ports)
- Avoid adding new deps unless justified

Tests
- Write tests first for use cases (fakes for ports)
- Integration tests under `test/integration` behind tags

Docs
- Update `README.md` and `docs/` for user-visible changes

