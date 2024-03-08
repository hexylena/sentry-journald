# Contributing

Contributions are welcome!

## Goals

- Small (LoC, not binary size, go makes such fat binaries but I don't really care.)
- Simple interface 

    - No filtering, UI configuration
    - Authentication should be handled upstream

- Simple to deploy

    - No database
    - No external dependencies
    - No configuration

## Anti-Goals

- Feature parity with Sentry
- Performance is nice but a secondary concern. (If it was a concern we wouldn't be using journald eh?)
