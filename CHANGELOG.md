## [0.7.4](https://github.com/imup-io/client/compare/v0.7.3...v0.7.4) (2023-01-02)


### Bug Fixes

* add link to godoc ([0672816](https://github.com/imup-io/client/commit/067281685dd40e0436f443b4decfb0c4d272c60d))

## [0.7.3](https://github.com/imup-io/client/compare/v0.7.2...v0.7.3) (2023-01-02)


### Bug Fixes

* do not generate release-notes for commits marked as chore ([862bc46](https://github.com/imup-io/client/commit/862bc464f7040e027d85bfee01edc82852f9ec87))

## [0.7.2](https://github.com/imup-io/client/compare/v0.7.1...v0.7.2) (2023-01-02)


### Bug Fixes

* use imup-bot as committer, not semantic-release-bot ([96cfb3c](https://github.com/imup-io/client/commit/96cfb3c06186fe156c326a1b2c57f4863e87fc55))

## [0.7.1](https://github.com/imup-io/client/compare/v0.7.0...v0.7.1) (2023-01-02)


### Bug Fixes

* **release:** enable changelog built with release notes ([75469e1](https://github.com/imup-io/client/commit/75469e1b92989f6b09752fec6d92d8df1fce296b))

# [0.7.0](https://github.com/imup-io/client/compare/v0.6.1...v0.7.0) (2023-01-02)


### Bug Fixes

* add changelog, ignore generated files during release process ([1b5847e](https://github.com/imup-io/client/commit/1b5847eb44e78815124a94d6e9cfb038dd078697))
* do not mutate id, key or email on reload ([8f003c1](https://github.com/imup-io/client/commit/8f003c10adece32fa721357542c4cca60650d0f3))
* ignore changelog ([fbc4b3a](https://github.com/imup-io/client/commit/fbc4b3ab5e68bfd30f54009f14d0eb369b5fd58e))
* lint ([f53ea8d](https://github.com/imup-io/client/commit/f53ea8dd692b71cc5be35fe478816c92100d3d2a))
* **release:** do not append to changelog, use the default ([c72b361](https://github.com/imup-io/client/commit/c72b36142c792c435e6d9de3149bb10a8ba4b017))
* **releaser:** [skip ci] set fetch-depth so releaser can find existing tags ([84593c5](https://github.com/imup-io/client/commit/84593c57516a6c80b4aa96d614863e54324761d3))
* **releaser:** add changelog and git release ([6567809](https://github.com/imup-io/client/commit/65678094197157cd0d450ee8d7a2df82b45d7ba2))
* **releaser:** add changelog config to goreleaser yaml ([c352061](https://github.com/imup-io/client/commit/c3520611661c39b81faa94276bf29e6ec01406c6))
* **releaser:** add changelog to gitignore ([0202c5b](https://github.com/imup-io/client/commit/0202c5be133530c8a257e0afc8256533e65afedd))
* **releaser:** do not pass release notes to goreleaser ([b49d585](https://github.com/imup-io/client/commit/b49d585d40fa4820ffde03430b5e66a13b013a37))
* **releaser:** remove before hook ([2a5305a](https://github.com/imup-io/client/commit/2a5305a39239a39c078a13551c6b316b650103a1))
* **releaser:** remove changelog action from semantic-release ([da5cdf1](https://github.com/imup-io/client/commit/da5cdf1f4d32f9fbcab098ed09fc1da1eae6d65d))
* **releaser:** use .Env map for env vars ([847e022](https://github.com/imup-io/client/commit/847e022aa564ab2b546d0c4a12fa6a63f62349bb))
* **releaser:** use default template for archive name ([a4cacf3](https://github.com/imup-io/client/commit/a4cacf39243cd973e1aeee6047856a2e52131983))
* reloadable config goroutine should use a cancellable context ([0b631f9](https://github.com/imup-io/client/commit/0b631f92303b3452ce6130126f1bd8a738311ab5))


### Features

* add group id,name some rearragement of existing vars ([67008a2](https://github.com/imup-io/client/commit/67008a2f2adb87b65ef64d3484e25ae6d3d50804))
* add group name and id to config reload requests ([f735be0](https://github.com/imup-io/client/commit/f735be0121e78a899a6cbe35946b62386609a401))
* create package utilities to share between packages ([da975cf](https://github.com/imup-io/client/commit/da975cf947150dc3b211f82241a872c5a77b6f00))
* release from ci ([16a0aa7](https://github.com/imup-io/client/commit/16a0aa77c1eb907de9462f447f4730b9ce090191))
* update client dependencies ([363420b](https://github.com/imup-io/client/commit/363420bd4d1d9e2004df526cf1df9662a9d03ef4))
