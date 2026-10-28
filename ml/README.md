# Observation contracts

Python scaffold for Smart Home Hub, days 11–15. `smarthome_ml.observations.Observation.parse`
validates versioned finite scalar observations with UTC timestamps and strict fields.
It does not open devices, load models, run DSP or serve an inference API.
`python -m smarthome_ml --version` works; a normal invocation exits 2 with NOT_IMPLEMENTED.

From repository root: `make test-python`; package build: `make build`.
Development tools are fixed in `requirements-dev.lock`; runtime dependencies are empty.
