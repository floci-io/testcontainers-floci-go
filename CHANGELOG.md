# Changelog

## [1.1.0](https://github.com/floci-io/testcontainers-floci-go/compare/v1.0.0...v1.1.0) (2026-10-10)


### Features

* shared core with a cloud descriptor; AWS module moves to flociaws ([#17](https://github.com/floci-io/testcontainers-floci-go/issues/17)) ([fbfa4ee](https://github.com/floci-io/testcontainers-floci-go/commit/fbfa4eecab0787b195e2c3cef7779e0d45020d8e))

## 1.0.0 (2026-10-07)


* feat!: Run follows the testcontainers-go module convention ([#12](https://github.com/floci-io/testcontainers-floci-go/issues/12)) ([a89718a](https://github.com/floci-io/testcontainers-floci-go/commit/a89718a84db3ce25e55442296c3b31334a23c584)), closes [#3](https://github.com/floci-io/testcontainers-floci-go/issues/3)


### Bug Fixes

* compilation issue ([97d479b](https://github.com/floci-io/testcontainers-floci-go/commit/97d479bbebdd32d2e967edb4d6d723c0287c8853))
* stop publishing ~420 service ports to the host by default ([#7](https://github.com/floci-io/testcontainers-floci-go/issues/7)) ([a1bf81d](https://github.com/floci-io/testcontainers-floci-go/commit/a1bf81db788b681526843f66eca7e16b32e4eb69))
* update to make testcontainers go standard ([#4](https://github.com/floci-io/testcontainers-floci-go/issues/4)) ([c2e1c03](https://github.com/floci-io/testcontainers-floci-go/commit/c2e1c031b558180a76e4e12a1c85503eaeb947ad))


### Features

* add DynamoDB/SQS/SNS/Lambda examples, CI matrix, and fix module tagging ([7c52b8f](https://github.com/floci-io/testcontainers-floci-go/commit/7c52b8f5dce636635239cb901eca9bc1221d3367))


### BREAKING CHANGES

* Run takes the image as its second argument and accepts
testcontainers.ContainerCustomizer options: floci.Run(ctx) becomes
floci.Run(ctx, "floci/floci:latest"), and closures become
floci.Option(...) or the package-level options. Stop(ctx) is removed
because it shadowed testcontainers.Container's Stop; use Terminate or
testcontainers.CleanupContainer.

Signed-off-by: Hector Ventura <hectorvent@gmail.com>
