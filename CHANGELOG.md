# [0.23.0](https://github.com/imup-io/client/compare/v0.22.0...v0.23.0) (2023-07-08)


### Bug Fixes

* include groupID in realtime data ([896f628](https://github.com/imup-io/client/commit/896f6282a8656407dc32ea573f07bc554e160cc4))


### Features

* include group id with speed and connectivity data ([f859016](https://github.com/imup-io/client/commit/f8590168d3879b72c53dc5fe8de92fdd7b9ea8c1))

# [0.22.0](https://github.com/imup-io/client/compare/v0.21.1...v0.22.0) (2023-07-08)


### Bug Fixes

* **goreleaser:** output binary name ([fb87dc5](https://github.com/imup-io/client/commit/fb87dc5f1195b97f381976efe24d1ea44c90178b))


### Features

* allow api server host override during build ([738a4ed](https://github.com/imup-io/client/commit/738a4edd29df633fd1c1fab105d13fd0fc516661))

## [0.21.1](https://github.com/imup-io/client/compare/v0.21.0...v0.21.1) (2023-07-08)


### Bug Fixes

* cidr warnings on default ping addresses ([1a02fc4](https://github.com/imup-io/client/commit/1a02fc4c464e3c9fcd764d3ddc6ebe8786480bf4))

# [0.21.0](https://github.com/imup-io/client/compare/v0.20.0...v0.21.0) (2023-07-08)


### Bug Fixes

* ordering ([d9cc293](https://github.com/imup-io/client/commit/d9cc293097fe38ff0cfd38ad966dac17cde8eb1b))


### Features

* update external ip settings to accept cider ranges ([5505f34](https://github.com/imup-io/client/commit/5505f3453add53eef0f18f7b730dffed4a818163))

# [0.20.0](https://github.com/imup-io/client/compare/v0.19.0...v0.20.0) (2023-07-07)


### Features

* refresh public IP every 1 minute instead of 15 ([f7bc085](https://github.com/imup-io/client/commit/f7bc08550dbe6a2b42170829816d81be17fa5c08))

# [0.19.0](https://github.com/imup-io/client/compare/v0.18.1...v0.19.0) (2023-07-03)


### Bug Fixes

* ci skipped test ci comments ([0b29312](https://github.com/imup-io/client/commit/0b29312ea28b46bc7478c15734c54c166f3532df))
* cleanup speed test realtime result ([ae1c03b](https://github.com/imup-io/client/commit/ae1c03bbeacba39abca9699022db7466000cc8a2))
* flaky test ([e27d1c2](https://github.com/imup-io/client/commit/e27d1c26c5bef0a0e20d5887ea4630d993b2c0de))
* **lint:** address lint issues discovered by golangci-lint ([b68fbad](https://github.com/imup-io/client/commit/b68fbadc5e80692348e5b0573f2cccc14ed09445))
* remove old comment ([97a38b3](https://github.com/imup-io/client/commit/97a38b341013f7c8393ebcb13cb28e6c6c9c72ad))
* update api tests ([ebf7b9a](https://github.com/imup-io/client/commit/ebf7b9a3bd16ec589d0bd83607a007ad87489cd0))


### Features

* add speed test options struct, allow a pass through server/url to override default locate api ([17eb4c3](https://github.com/imup-io/client/commit/17eb4c32b6e60748273a5a378336e2e7d40d990a))
* allow the speed test backoff function to be cancellable ([ddc4697](https://github.com/imup-io/client/commit/ddc46977e6d00400904298bc4a1a2ef4c46b2575))
* CI ([789ff7c](https://github.com/imup-io/client/commit/789ff7c86eee3c6aa7c85fe4d77fd96f4c99d744))
* move speed testing into its own package ([7009068](https://github.com/imup-io/client/commit/7009068ebf5027ad4a417aa18efe5317f153c660))
* refactor speedtesting ([3dbd8f1](https://github.com/imup-io/client/commit/3dbd8f14d6c105fdd44d7e21abc7e38e10c486c4))
* update tests to use a mock speed test server ([d7e649e](https://github.com/imup-io/client/commit/d7e649ec2424796758c8e60b2a3579c0e8d850d1))
* use new speed testing package ([bcce709](https://github.com/imup-io/client/commit/bcce7091b64fd744f757e9eb111738136e609b95))
* use package speedtesting ([175d375](https://github.com/imup-io/client/commit/175d375d315254b2b62c2528a038d403ed6dbcb4))

## [0.18.1](https://github.com/imup-io/client/compare/v0.18.0...v0.18.1) (2023-06-04)


### Bug Fixes

* make the json tag for Group consistent ([761aebf](https://github.com/imup-io/client/commit/761aebf62f3e2dac8c4c186abf3cd42a67d1d6e7))

# [0.18.0](https://github.com/imup-io/client/compare/v0.17.0...v0.18.0) (2023-05-31)


### Bug Fixes

* do not expose avoid addrs ([12bd66b](https://github.com/imup-io/client/commit/12bd66b251108b41fca7b9af5bb16f7a05780e82))
* remove unused ([7f0316e](https://github.com/imup-io/client/commit/7f0316edc70e7794dcf5e8b5bfcc0cdc2a73d261))


### Features

* add new interface and types for connectivity data ([fe37e69](https://github.com/imup-io/client/commit/fe37e692d9a0994d487bc1b867c0035ecceb8dac))
* move connectivity testing into its own folder ([2e7bed6](https://github.com/imup-io/client/commit/2e7bed6bcba3ac7035aaac68ac95b0dd2d209003))
* options, not connectivity options ([2861829](https://github.com/imup-io/client/commit/2861829b4704e99c277e096eb1c30fc66854c17b))
* update dial to be more modular ([7a464dd](https://github.com/imup-io/client/commit/7a464dd169eb6b41b6c67cd738170fd76e537679))
* update ping to be more modular ([c8e7b81](https://github.com/imup-io/client/commit/c8e7b81d5f426a2bbe91aad014651841a46e400e))
* use package connectivity ([db479b9](https://github.com/imup-io/client/commit/db479b96aad70df46de5d8e0cd77cd8758071e5a))

# [0.17.0](https://github.com/imup-io/client/compare/v0.16.0...v0.17.0) (2023-05-06)


### Bug Fixes

* keep GroupID as the public interface to a reloadable config ([ee94652](https://github.com/imup-io/client/commit/ee946525fa6b05581c152254adc9de98b7588f60))
* update flag descriptors ([7aa27bc](https://github.com/imup-io/client/commit/7aa27bc18506b56db93984f8ca20451020857af5))
* update internal references to imup config ([cd0227e](https://github.com/imup-io/client/commit/cd0227eb4966069189ced7983edc312df9c5ef51))


### Features

* move all of the configurable bits into config ([5530750](https://github.com/imup-io/client/commit/5530750e591ef22b39f4f1c18f7c7063ae95e626))
* organize and fix naming of existing config as is ([0008f43](https://github.com/imup-io/client/commit/0008f434453384258b0f67a627bdb7ab3112b9f5))
* ping is now enabled by default ([6f4596a](https://github.com/imup-io/client/commit/6f4596a9853a9050298c87bae153a3bc39d1cec6))

# [0.16.0](https://github.com/imup-io/client/compare/v0.15.2...v0.16.0) (2023-05-05)


### Bug Fixes

* do not swap logger out while config is locked ([cab7d57](https://github.com/imup-io/client/commit/cab7d570753c554153404fb1abb6fb93d6d583d8))
* pr comments ([1e07ae5](https://github.com/imup-io/client/commit/1e07ae5e2930da0f024e50d9b0c666c7d99cf3c9))
* remove emitter output, fix bug in speed tests on windows where streaming to stderr caused the tests to stop early ([29303ac](https://github.com/imup-io/client/commit/29303acb60e5e55f38121fc90eb838ad09b7eeb4))
* run a mod tidy after installing go ([053ba05](https://github.com/imup-io/client/commit/053ba05b3937919bb0d0889871ae02d9a1b96353))
* separate reloadble functions into their own file ([a7be59e](https://github.com/imup-io/client/commit/a7be59e842279d5175bcc900baced1a25ae6b26f))
* try a specific version of go ([4c001df](https://github.com/imup-io/client/commit/4c001dff4733562bab780e0a9d01893217e39bb5))
* update ci to use go120 ([c868969](https://github.com/imup-io/client/commit/c86896984855395c081471ecf69e93e8d0a6a204))
* use correct version of staticcheck ([7b9cf04](https://github.com/imup-io/client/commit/7b9cf042ef8acb0dd4666c66a0e0d2c8de3f320b))
* use setup-go action ([2b8bfed](https://github.com/imup-io/client/commit/2b8bfed7824152236b559af4a46e305356aea0ae))


### Features

* add support for running imup as a windows service ([c404779](https://github.com/imup-io/client/commit/c404779e3b18f5438dd4813dd9145085d9ddd59d))
* configurable log verbosity and log to file ([7a2d145](https://github.com/imup-io/client/commit/7a2d145dff52e7e49cb48a391feb0b54f4d4d097))
* reloadable logger ([4cb3c73](https://github.com/imup-io/client/commit/4cb3c73db9e54899bab97c1031e381ee34770599))

## [0.15.2](https://github.com/imup-io/client/compare/v0.15.1...v0.15.2) (2023-04-21)


### Bug Fixes

* boolean pointers have a default value of false always ([c819ba6](https://github.com/imup-io/client/commit/c819ba6f726d5a034941b71165c8b6f15a819654))
* ensure passed in flag value takes precedence over an environment variable ([5ecf4f6](https://github.com/imup-io/client/commit/5ecf4f69ef0c36703ed3ecffee1dc202b9a5cc37))

## [0.15.1](https://github.com/imup-io/client/compare/v0.15.0...v0.15.1) (2023-04-15)


### Bug Fixes

* archives need an id ([4d32a95](https://github.com/imup-io/client/commit/4d32a952bed0d4e94ebff129edd01798cab5347e))

# [0.15.0](https://github.com/imup-io/client/compare/v0.14.0...v0.15.0) (2023-04-15)


### Features

* configure windows releaser to use a zip archive ([cc26592](https://github.com/imup-io/client/commit/cc265922d533580c9b2bb6ef9cab2f4063bbdaf0))

# [0.14.0](https://github.com/imup-io/client/compare/v0.13.0...v0.14.0) (2023-04-09)


### Bug Fixes

* do not camelCase flags ([58caf21](https://github.com/imup-io/client/commit/58caf21c80332b548b5b960d0d4e3dc05fc12994))
* do not export read-only configuration ([6aeda67](https://github.com/imup-io/client/commit/6aeda673f95bcc1bf790e93fdab205f7d8ba3cd5))
* multiple imports of slog ([625dc83](https://github.com/imup-io/client/commit/625dc839c292034ccbe4437720f3b25761d0eefa))


### Features

* add PublicIP and RefreshPublicIP functions to config ([28b9c24](https://github.com/imup-io/client/commit/28b9c243bf0679d5cfbbbc8eaa720fa2afaf1009))
* fetch a clients public ip address on startup ([bee6bf3](https://github.com/imup-io/client/commit/bee6bf3b2bd63b36c9c5cc890aae613ee9d1c7f0))
* log new configs ([3e413ed](https://github.com/imup-io/client/commit/3e413ed91f8e2f92be10ec164c4cb18865d52847))
* realtime first, use public IP from config, remove cache ([5be4aaa](https://github.com/imup-io/client/commit/5be4aaac08a44a0cfead7643533f7ab40e97378a))

# [0.13.0](https://github.com/imup-io/client/compare/v0.12.0...v0.13.0) (2023-04-03)


### Bug Fixes

* suppress warnings when ip is empty, fix coverage by clearing env ([3769796](https://github.com/imup-io/client/commit/3769796c8a9424d1efc6f48de6e3a6abafe1c238))
* use the correct env var for setting allow/block listed ips ([6fae638](https://github.com/imup-io/client/commit/6fae638bd84351562f724f890f6310cc85c9c9b9))


### Features

* sicne we do not allow for mutating a clients existing id,key or email ensure the existing config is part of the new config ([78ae103](https://github.com/imup-io/client/commit/78ae103a93015b2df2fc6d518e86e3c669e58737))

# [0.12.0](https://github.com/imup-io/client/compare/v0.11.0...v0.12.0) (2023-04-01)


### Bug Fixes

* update dockerfile to match current go version ([3ce3c74](https://github.com/imup-io/client/commit/3ce3c740fce4b9dcef2de71417b879c69fd71ac5))
* use the unified ipify endpoint to return ipv6 as well as ipv4 addresses ([c8d3a42](https://github.com/imup-io/client/commit/c8d3a4261f57f073dd5073dd44856c0bfa6d1c7b))


### Features

* add unit test for mixing cidr range and single ip ([79c6319](https://github.com/imup-io/client/commit/79c6319c6b3596a7d92ed9815ee0143d2a03b795))
* extend allow/block listed ips to consider cidr ranges ([3edfc8a](https://github.com/imup-io/client/commit/3edfc8a064879bca96c1254796a2e0d00511bbb9))

# [0.11.0](https://github.com/imup-io/client/compare/v0.10.0...v0.11.0) (2023-04-01)


### Features

* add groug-id and name to speed and connectivity data ([22feb82](https://github.com/imup-io/client/commit/22feb826aef946d53ce3721aaa4a2528cc6078ee))

# [0.10.0](https://github.com/imup-io/client/compare/v0.9.0...v0.10.0) (2023-04-01)


### Bug Fixes

* additional context for error messages ([5d5cc66](https://github.com/imup-io/client/commit/5d5cc66c09fee633fb6d54726bafa942d07bcec6))
* rand.Seed deprecated as auto seeding is now the default https://github.com/golang/go/issues/54880 ([f3127d0](https://github.com/imup-io/client/commit/f3127d079e1c3c97f8e377565239a1b9ad69d597))
* use log not slog ([591fe0c](https://github.com/imup-io/client/commit/591fe0ca2a9e46850cea399144308fb860e27b56))


### Features

* replace logrus with slog ([e67098f](https://github.com/imup-io/client/commit/e67098f9c9fb9030f77a7df08e84c171ed5c724d))

# [0.9.0](https://github.com/imup-io/client/compare/v0.8.2...v0.9.0) (2023-03-17)


### Bug Fixes

* install go in releaser action (do not use system go) ([2c0f910](https://github.com/imup-io/client/commit/2c0f910631ebd8a56ededc687daaf27196a19fe1))


### Features

* update deps and language to latest ([5894a0c](https://github.com/imup-io/client/commit/5894a0cee0160d8b0f52fcebd6eb9e54e0f2ecf3))

## [0.8.2](https://github.com/imup-io/client/compare/v0.8.1...v0.8.2) (2023-03-17)


### Bug Fixes

* remove trailing slash ([f8d3a0c](https://github.com/imup-io/client/commit/f8d3a0c802c78175709331c42a4b91dc3da31cd1))

## [0.8.1](https://github.com/imup-io/client/compare/v0.8.0...v0.8.1) (2023-03-17)


### Bug Fixes

* space in address ([1b532c3](https://github.com/imup-io/client/commit/1b532c362afd5533f71e24516758a2de33508bf9))

# [0.8.0](https://github.com/imup-io/client/compare/v0.7.4...v0.8.0) (2023-03-17)


### Bug Fixes

* add block/allow ips in the same test ([6f007b2](https://github.com/imup-io/client/commit/6f007b2a11c08e30948fe35692cf3a0e3ad322a3))
* dependabot-go_modules-github.com-hashicorp-go-retryablehttp-0.7.2 ([330f9ba](https://github.com/imup-io/client/commit/330f9baefaabcc175890f1ba17aa126514d6d545))
* dependabot-go_modules-github.com-matryer-is-1.4.1 ([5eb76ee](https://github.com/imup-io/client/commit/5eb76ee23627fafa873d6236c9e6f53ec6854cfb))
* go mod tidy ([ca47799](https://github.com/imup-io/client/commit/ca47799ce5cfc72d8c191d4202897015dfd9a80d))
* unused variable ([9b38d83](https://github.com/imup-io/client/commit/9b38d83656391a581147ef4db2fc508e9a64805e))


### Features

* extend ip allowing/blocking to connectivity testing ([d142f71](https://github.com/imup-io/client/commit/d142f71e75872f3c3b988516c30b5094905c8222))
* implement allow/block lists for speed testing ([4b9cf29](https://github.com/imup-io/client/commit/4b9cf29cf83567f1bf4168a7197652c304f26f1b))

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
