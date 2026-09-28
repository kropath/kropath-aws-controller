# Changelog

## [0.15.0](https://github.com/kropath/kropath-controller/compare/v0.14.0...v0.15.0) (2026-09-28)


### Features

* **KRO-1258:** add namingTemplate to Route53Config cascade ([#150](https://github.com/kropath/kropath-controller/issues/150)) ([d91ec74](https://github.com/kropath/kropath-controller/commit/d91ec74e4df14161253e3f9d5141557d0856f33e))
* **KRO-1278:** metrics foundation, collector pattern, non-leader runnable ([#159](https://github.com/kropath/kropath-controller/issues/159)) ([6de077e](https://github.com/kropath/kropath-controller/commit/6de077e4fc1e3def679a73eadb3069f1e225c822))
* **KRO-1279:** metrics for config cascade, KropathConfigStatus, NamespacePlacement ([#161](https://github.com/kropath/kropath-controller/issues/161)) ([6657bf3](https://github.com/kropath/kropath-controller/commit/6657bf39d0155dbfa0a1aaf1d22626c8046832a4))
* **KRO-1280:** policydocument observability metrics + M-10 collect-all ([#160](https://github.com/kropath/kropath-controller/issues/160)) ([41d2e07](https://github.com/kropath/kropath-controller/commit/41d2e0773f297745b5fe865a03e56a2bc4f72159))
* **KRO-1281:** label injection + registry/dynamic-CRD metrics ([#162](https://github.com/kropath/kropath-controller/issues/162)) ([66c87a3](https://github.com/kropath/kropath-controller/commit/66c87a3cebed5797f8f8f919272091d2dea12052))
* **KRO-1282:** monitoring alerting rules, label-value CI check, and metrics docs ([#163](https://github.com/kropath/kropath-controller/issues/163)) ([51cd633](https://github.com/kropath/kropath-controller/commit/51cd63388201bdee6769fd6b5731a3aeaada3c1f))
* **KRO-1283:** grafana dashboard and non-leader collector Chainsaw suite ([#165](https://github.com/kropath/kropath-controller/issues/165)) ([f5b236f](https://github.com/kropath/kropath-controller/commit/f5b236f627601a752a66650cf953636a2cee5b7c))


### Bug Fixes

* **KRO-1263:** wire 21 orphaned Chainsaw suites into test-chainsaw ([#152](https://github.com/kropath/kropath-controller/issues/152)) ([5c7c8c5](https://github.com/kropath/kropath-controller/commit/5c7c8c580fc677e99e60dd2dffdbecfe5464f2b2))


### Dependencies

* bump the kubernetes group across 1 directory with 3 updates ([#164](https://github.com/kropath/kropath-controller/issues/164)) ([8fd946a](https://github.com/kropath/kropath-controller/commit/8fd946af87de7945a47bb6fc19c4fc75fc06903c))

## [0.14.0](https://github.com/kropath/kropath-controller/compare/v0.13.0...v0.14.0) (2026-09-26)


### Features

* **KRO-1228:** extend CloudFront effectiveConfig cascade with vpcOriginProtocolPolicy ([#143](https://github.com/kropath/kropath-controller/issues/143)) ([38e663c](https://github.com/kropath/kropath-controller/commit/38e663caa17d4592171d993794d1ba70375c9d82))
* **KRO-1234:** extend KMSConfig cascade with allowedGrantOperations ([#144](https://github.com/kropath/kropath-controller/issues/144)) ([eae9b5f](https://github.com/kropath/kropath-controller/commit/eae9b5fbf4ed5010ad4144e94909a1b9e63bc4cc))


### Bug Fixes

* **KRO-1241:** remove dead KropathConfig.spec.aws from CRD fixture and Chainsaw setups ([#146](https://github.com/kropath/kropath-controller/issues/146)) ([d2b9413](https://github.com/kropath/kropath-controller/commit/d2b94137596be9abecc52cda018834803770f09e))

## [0.13.0](https://github.com/kropath/kropath-controller/compare/v0.12.0...v0.13.0) (2026-09-22)


### Features

* **KRO-1128:** resolve account/region placement from namespace annotations ([#140](https://github.com/kropath/kropath-controller/issues/140)) ([8c79d1e](https://github.com/kropath/kropath-controller/commit/8c79d1e63d50215b40b594174250e6d22ca6c025))
* **KRO-1141:** add ACK install conformance checker ([#124](https://github.com/kropath/kropath-controller/issues/124)) ([c25a703](https://github.com/kropath/kropath-controller/commit/c25a70316dfa510607b444e36e6553f5a61e06b2))


### Bug Fixes

* **KRO-1199:** make status.effectiveConfig genuinely absent when withheld ([#142](https://github.com/kropath/kropath-controller/issues/142)) ([72fdd7c](https://github.com/kropath/kropath-controller/commit/72fdd7c578e632578da1904336fed27a72276d13))


### Dependencies

* bump sigs.k8s.io/controller-runtime in the kubernetes group ([#141](https://github.com/kropath/kropath-controller/issues/141)) ([e8e1db4](https://github.com/kropath/kropath-controller/commit/e8e1db410aa4acdd2fc260f0f9eeb307d41881b3))

## [0.12.0](https://github.com/kropath/kropath-controller/compare/v0.11.0...v0.12.0) (2026-09-19)


### Features

* **KRO-1119:** enforce singleton KropathConfig name and fixed global lookup ([#122](https://github.com/kropath/kropath-controller/issues/122)) ([096076b](https://github.com/kropath/kropath-controller/commit/096076b4190bbca19382ce7aab0217daf4aa4200))
* **KRO-1120:** implement ADR-015 §5.9 global-tier profile fallthrough ([#121](https://github.com/kropath/kropath-controller/issues/121)) ([7e08756](https://github.com/kropath/kropath-controller/commit/7e0875690ea51f6da62fc154b904d36fa59a4a0b))
* **KRO-1121:** publish Reconciled status condition on KropathConfig ([#126](https://github.com/kropath/kropath-controller/issues/126)) ([fe6b602](https://github.com/kropath/kropath-controller/commit/fe6b6028fdddc4965b1053415764dce90e872d80))


### Bug Fixes

* **KRO-1137:** apply Option A namespace-pair migration to 3 more suites ([#130](https://github.com/kropath/kropath-controller/issues/130)) ([bcbd6c2](https://github.com/kropath/kropath-controller/commit/bcbd6c2e3c98584fc3354dceb20aaee5bdc75bd1))
* **KRO-1137:** apply Option A namespace-pair migration to 4 more suites ([#134](https://github.com/kropath/kropath-controller/issues/134)) ([a8ec2d5](https://github.com/kropath/kropath-controller/commit/a8ec2d532a39f218f6407ba6fe22387b3d419f98))
* **KRO-1137:** apply Option A namespace-pair migration to 5 more suites ([#132](https://github.com/kropath/kropath-controller/issues/132)) ([04d31c0](https://github.com/kropath/kropath-controller/commit/04d31c0327d52037a3e7e3ab2bd5bb1975069e23))
* **KRO-1137:** apply Option A namespace-pair migration to 5 more suites ([#133](https://github.com/kropath/kropath-controller/issues/133)) ([45be7fe](https://github.com/kropath/kropath-controller/commit/45be7fef2b63753b392536e65dea4ece20bfa0ab))
* **KRO-1137:** apply Option A namespace-pair migration to 6 more suites ([#128](https://github.com/kropath/kropath-controller/issues/128)) ([282fdd2](https://github.com/kropath/kropath-controller/commit/282fdd29707d51f767e4824382ac365e1dc7667c))
* **KRO-1137:** apply Option A namespace-pair migration to 6 more suites ([#129](https://github.com/kropath/kropath-controller/issues/129)) ([bea1cd5](https://github.com/kropath/kropath-controller/commit/bea1cd54f0798f6372bc7904861b7a18bda26c83))
* **KRO-1137:** apply Option A namespace-pair migration to ctrl-acm, ctrl-dsql-01 ([#127](https://github.com/kropath/kropath-controller/issues/127)) ([3831558](https://github.com/kropath/kropath-controller/commit/38315584a61c2a90d9022d7c3353ca772066c321))
* **KRO-1137:** apply suite-wide namespace-pair migration to the last 5 numbered-file suites ([#135](https://github.com/kropath/kropath-controller/issues/135)) ([42a7f72](https://github.com/kropath/kropath-controller/commit/42a7f724808acf060ba9bb4502d7c969e46dec15))
* **KRO-1137:** fix migration script header bug + apply namespace-pair migration to 4 more suites ([#131](https://github.com/kropath/kropath-controller/issues/131)) ([30b4b4b](https://github.com/kropath/kropath-controller/commit/30b4b4bc619d80b050726cb376a028b6ca6768e9))
* **KRO-1137:** rename KropathConfig fixtures to baseline in 7 verified-safe suites ([#125](https://github.com/kropath/kropath-controller/issues/125)) ([5b1d2cd](https://github.com/kropath/kropath-controller/commit/5b1d2cdc75afb863c6ac6a6c3222bd05b18b9191))
* **KRO-1149:** sync KropathConfig test fixture, rename singletons, set companion CRD ref ([#137](https://github.com/kropath/kropath-controller/issues/137)) ([b0ada80](https://github.com/kropath/kropath-controller/commit/b0ada804134830660409b3833cb208723cf2b982))
* **KRO-1150:** add GlobalProfileResolved condition and baseline singleton naming to bedrock/sagemaker/lambda asserts ([#136](https://github.com/kropath/kropath-controller/issues/136)) ([6fd1e53](https://github.com/kropath/kropath-controller/commit/6fd1e537502066ac713a3705e110a40275582323))

## [0.11.0](https://github.com/kropath/kropath-controller/compare/v0.10.0...v0.11.0) (2026-09-17)


### Features

* **KRO-1096:** wire S3AdvancedConfig cascade reconciler and controller-origin CRD ([b9a3149](https://github.com/kropath/kropath-controller/commit/b9a31490e9089f81a2e4ed595cc22c458e389eae))
* **KRO-1097:** add CloudFrontConfig cascade reconciler and controller-origin CRD ([62b5b34](https://github.com/kropath/kropath-controller/commit/62b5b34a30376ab59135970e311acaa323fbbaad))
* **KRO-1098:** add LambdaConfig effectiveConfig cascade reconciler ([#120](https://github.com/kropath/kropath-controller/issues/120)) ([145ec07](https://github.com/kropath/kropath-controller/commit/145ec073303cb61b348b83d8444376e0ee640d7b))
* **KRO-1100:** extend crds-verify with spec-default drift and kind-set drift checks ([#116](https://github.com/kropath/kropath-controller/issues/116)) ([927aaaf](https://github.com/kropath/kropath-controller/commit/927aaafe94e0490b361ccdbb0dc48352eb59795f))

## [0.10.0](https://github.com/kropath/kropath-controller/compare/v0.9.2...v0.10.0) (2026-09-14)


### Features

* **KRO-1081:** effectiveConfig cascade for S3 Advanced fields ([08f6ab6](https://github.com/kropath/kropath-controller/commit/08f6ab685220e6f89c5fb830bb6057953bcbcd3a))

## [0.9.2](https://github.com/kropath/kropath-controller/compare/v0.9.1...v0.9.2) (2026-09-12)


### Bug Fixes

* **KRO-851:** declare all eight watched kinds for PolicyDocument in features.All ([#112](https://github.com/kropath/kropath-controller/issues/112)) ([81bd624](https://github.com/kropath/kropath-controller/commit/81bd62468ad72d36b83f0be4a6d9b459a570d416))

## [0.9.1](https://github.com/kropath/kropath-controller/compare/v0.9.0...v0.9.1) (2026-09-11)


### Dependencies

* bump github.com/prometheus/client_golang from 1.24.0 to 1.24.1 ([#104](https://github.com/kropath/kropath-controller/issues/104)) ([3bf4c05](https://github.com/kropath/kropath-controller/commit/3bf4c05b44c6ef837afdd9f189584804b40929a8))
* bump the gomod-patch group with 2 updates ([#110](https://github.com/kropath/kropath-controller/issues/110)) ([edddb1e](https://github.com/kropath/kropath-controller/commit/edddb1e9b977fd0bf502063d75fcdadee3bcfadc))
* bump the kubernetes group across 1 directory with 3 updates ([#103](https://github.com/kropath/kropath-controller/issues/103)) ([56ff281](https://github.com/kropath/kropath-controller/commit/56ff2817858f6b5aea3896ee84817d2d7d009f05))

## [0.9.0](https://github.com/kropath/kropath-controller/compare/v0.8.0...v0.9.0) (2026-09-09)


### Features

* **KRO-1007:** add RAM effectiveConfig cascade controller ([#97](https://github.com/kropath/kropath-controller/issues/97)) ([7a8b455](https://github.com/kropath/kropath-controller/commit/7a8b455c18b3d5b3b9d138487a48f694019c571b))
* **KRO-1015:** add BackupConfig effectiveConfig cascade ([#94](https://github.com/kropath/kropath-controller/issues/94)) ([9db0b88](https://github.com/kropath/kropath-controller/commit/9db0b887cc1be00c841fe07188a9b5d0c9f8763a))
* **KRO-1023:** add QuickSight effectiveConfig cascade ([#98](https://github.com/kropath/kropath-controller/issues/98)) ([4cabc3b](https://github.com/kropath/kropath-controller/commit/4cabc3b63bdcfbfbe489fa46a6e9748552f381a8))
* **KRO-1031:** add ECRPublicConfig effectiveConfig cascade ([#99](https://github.com/kropath/kropath-controller/issues/99)) ([babd88f](https://github.com/kropath/kropath-controller/commit/babd88f7aba47141993f0b3775037fc5dcc8442c))
* **KRO-1039:** add EBS Recycle Bin effectiveConfig cascade ([#100](https://github.com/kropath/kropath-controller/issues/100)) ([9020e6d](https://github.com/kropath/kropath-controller/commit/9020e6de62e84c73fce3a458007a6a751320007b))
* **KRO-866:** add MQConfig effectiveConfig cascade ([#96](https://github.com/kropath/kropath-controller/issues/96)) ([6b409c9](https://github.com/kropath/kropath-controller/commit/6b409c97a311b172040241be21e28b0653f6c4ed))
* **KRO-873:** effectiveConfig cascade for EventBridge Pipes fields ([8647822](https://github.com/kropath/kropath-controller/commit/8647822b583335e43342da405e0807f67f0a6f8d))
* **KRO-944:** effectiveConfig cascade: Network Firewall fields ([#93](https://github.com/kropath/kropath-controller/issues/93)) ([157b3f9](https://github.com/kropath/kropath-controller/commit/157b3f9b8530c22579d2ea363f723a09c6e9f6b7))
* **KRO-953:** add ManagedPrometheus effectiveConfig cascade ([#95](https://github.com/kropath/kropath-controller/issues/95)) ([b66f3bd](https://github.com/kropath/kropath-controller/commit/b66f3bd35a434f6cf4c89bd2a84fb2e8f55d8996))
* **KRO-961:** add SESConfig effectiveConfig cascade ([#89](https://github.com/kropath/kropath-controller/issues/89)) ([9894516](https://github.com/kropath/kropath-controller/commit/98945166590db70e34c7f7517f097f0c21aa8fcf))
* **KRO-973:** add CodeArtifact effectiveConfig cascade support ([dd84566](https://github.com/kropath/kropath-controller/commit/dd845662a7ce9bdf157d9e140c0a0bc456595eaf))
* **KRO-979:** add LoggingTargetPrefix to S3 cascade ([#87](https://github.com/kropath/kropath-controller/issues/87)) ([ffb1900](https://github.com/kropath/kropath-controller/commit/ffb1900131a03e5da2e35f8eda851afbd41f94b7))
* **KRO-991:** add MWAA effectiveConfig cascade ([15a2c58](https://github.com/kropath/kropath-controller/commit/15a2c582b7a920fe1bc9802c1a778ed8bfc93c5d))
* **KRO-999:** effectiveConfig cascade for Organizations fields ([bc49dc9](https://github.com/kropath/kropath-controller/commit/bc49dc93f7bcb2c0a440226de5bc4a3828c04d9f))

## [0.8.0](https://github.com/kropath/kropath-controller/compare/v0.7.0...v0.8.0) (2026-09-02)


### Features

* **KRO-737:** implement CognitoConfig effectiveConfig cascade ([#76](https://github.com/kropath/kropath-controller/issues/76)) ([3038c6f](https://github.com/kropath/kropath-controller/commit/3038c6f61819b2c40d6fb4efabdb112ec302b5e4))
* **KRO-746:** effectiveConfig cascade for Kinesis fields ([#78](https://github.com/kropath/kropath-controller/issues/78)) ([2c26af9](https://github.com/kropath/kropath-controller/commit/2c26af95ca27620c700f0e3935db80f4e24d29a0))
* **KRO-762:** add CloudTrail effectiveConfig cascade ([#79](https://github.com/kropath/kropath-controller/issues/79)) ([5c89a92](https://github.com/kropath/kropath-controller/commit/5c89a92c7ae31b964c8832fa61f1c3e2874e9cea))
* **KRO-772:** add AppScalingConfig effectiveConfig cascade ([#80](https://github.com/kropath/kropath-controller/issues/80)) ([6e3d0cb](https://github.com/kropath/kropath-controller/commit/6e3d0cb2b32bad24e803a8111c080948e0cdfdb8))
* **KRO-790:** add Keyspaces effectiveConfig cascade ([9892c6c](https://github.com/kropath/kropath-controller/commit/9892c6c3a409144c23b575fc76b7ca127109e1a9))
* **KRO-806:** add OpenSearch effectiveConfig cascade controller ([#85](https://github.com/kropath/kropath-controller/issues/85)) ([a72a22e](https://github.com/kropath/kropath-controller/commit/a72a22e423bfc0591266c8337a060fe843d2c928))
* **KRO-814:** add Bedrock effectiveConfig cascade ([#82](https://github.com/kropath/kropath-controller/issues/82)) ([34e5af7](https://github.com/kropath/kropath-controller/commit/34e5af795161ffbf8bec5f112f73c88225cac2a3))
* **KRO-822:** add SageMakerConfig effectiveConfig cascade ([#84](https://github.com/kropath/kropath-controller/issues/84)) ([d5b02aa](https://github.com/kropath/kropath-controller/commit/d5b02aae1a9688ae97c5b697136d9a0d47c91071))
* **KRO-882:** add WAF effectiveConfig cascade ([#83](https://github.com/kropath/kropath-controller/issues/83)) ([b2b3680](https://github.com/kropath/kropath-controller/commit/b2b3680aeba907be3b39512acd63c55a1b1f2624))

## [0.7.0](https://github.com/kropath/kropath-controller/compare/v0.6.0...v0.7.0) (2026-08-28)


### Features

* **KRO-728:** add Route53Config effectiveConfig cascade ([#73](https://github.com/kropath/kropath-controller/issues/73)) ([b7c4e1f](https://github.com/kropath/kropath-controller/commit/b7c4e1f6bbac6d907e5dbb8c947eed41df091d8b))
* **KRO-754:** add SSMConfig effectiveConfig cascade ([#75](https://github.com/kropath/kropath-controller/issues/75)) ([7044d86](https://github.com/kropath/kropath-controller/commit/7044d86644c7c5e5829df7dbfcd1cfea9e019b08))

## [0.6.0](https://github.com/kropath/kropath-controller/compare/v0.5.0...v0.6.0) (2026-08-27)


### Features

* **KRO-798:** add DSQLConfig effectiveConfig cascade handler ([#72](https://github.com/kropath/kropath-controller/issues/72)) ([57b057e](https://github.com/kropath/kropath-controller/commit/57b057eba91204b5258c74c55ee67ca9e9aac6cf))
* **KRO-857:** effectiveConfig cascade for Athena fields ([ae7ea6c](https://github.com/kropath/kropath-controller/commit/ae7ea6cbb3b07f2f2865e0d2af49e3752c55f974))


### Bug Fixes

* **KRO-894:** pass --metrics-bind-address to chainsaw-start, bind both ports to loopback ([#70](https://github.com/kropath/kropath-controller/issues/70)) ([87b09ed](https://github.com/kropath/kropath-controller/commit/87b09ed6d90e8762b8ba6435d48fef3f359792ef))

## [0.5.0](https://github.com/kropath/kropath-controller/compare/v0.4.0...v0.5.0) (2026-08-24)


### Features

* **KRO-782:** add DocumentDB effectiveConfig cascade ([#64](https://github.com/kropath/kropath-controller/issues/64)) ([d39d78d](https://github.com/kropath/kropath-controller/commit/d39d78dc0924bd01f4b9a66620dd65d89587e3a5))
* **KRO-842:** add GlueConfig effectiveConfig cascade ([#68](https://github.com/kropath/kropath-controller/issues/68)) ([7c2293e](https://github.com/kropath/kropath-controller/commit/7c2293eaff30d43061c54266565162e61cd80716))
* **KRO-849:** add CRD watcher, features active-pending contract, RBAC, and Chainsaw ctrl-dyn-02 and ctrl-dyn-03 ([#66](https://github.com/kropath/kropath-controller/issues/66)) ([70769c9](https://github.com/kropath/kropath-controller/commit/70769c9205f5f11381bb062ca7e1a4d2669ad3e1))
* **KRO-850:** labeloperator convergence + policydocument optional kinds + metrics + Chainsaw ctrl-dyn-04/05 ([#67](https://github.com/kropath/kropath-controller/issues/67)) ([157093e](https://github.com/kropath/kropath-controller/commit/157093ef67583b61734513fadf468dd7f28f44c0))

## [0.4.0](https://github.com/kropath/kropath-controller/compare/v0.3.1...v0.4.0) (2026-08-24)


### Features

* **KRO-716:** add Certificate Manager effectiveConfig cascade ([#59](https://github.com/kropath/kropath-controller/issues/59)) ([76b5552](https://github.com/kropath/kropath-controller/commit/76b5552dfaa87ad350bbeca27f5a0f5efda9b1ea))
* **KRO-830:** implement effectiveConfig cascade handler for EMR fields ([#61](https://github.com/kropath/kropath-controller/issues/61)) ([6d25eb3](https://github.com/kropath/kropath-controller/commit/6d25eb3e8a8efe51e867b9e74e8339937c72805f))
* **KRO-848:** internal/registry + startup discovery gate ([#63](https://github.com/kropath/kropath-controller/issues/63)) ([d018bbb](https://github.com/kropath/kropath-controller/commit/d018bbb6f8d9f09ddfcdfefd738cf2b59fe1637e))
* **KRO-860:** add two-directory CRD fixture support (crds-optional/) ([#62](https://github.com/kropath/kropath-controller/issues/62)) ([bcbade2](https://github.com/kropath/kropath-controller/commit/bcbade22cc9ed6aef09b1bd26d6fa9119e577e33))

## [0.3.1](https://github.com/kropath/kropath-controller/compare/v0.3.0...v0.3.1) (2026-08-19)


### Bug Fixes

* **KRO-675:** match APIGatewayConfig kind casing to the kropath-aws CRD ([#56](https://github.com/kropath/kropath-controller/issues/56)) ([37756b1](https://github.com/kropath/kropath-controller/commit/37756b1def7b83eb94a53cfad17815363d509de9))

## [0.3.0](https://github.com/kropath/kropath-controller/compare/v0.2.0...v0.3.0) (2026-08-17)


### Features

* **KRO-348:** CloudWatch effectiveConfig cascade in kropath-controller ([a2dbc91](https://github.com/kropath/kropath-controller/commit/a2dbc912dfcee01fe34520c60331fd0e6625c4b8))
* **KRO-430:** implement MemoryDB effectiveConfig cascade ([#55](https://github.com/kropath/kropath-controller/issues/55)) ([e241328](https://github.com/kropath/kropath-controller/commit/e24132800d01a770fc399647f237468632d01bfa))
* **KRO-571:** add MSK effectiveConfig cascade ([#54](https://github.com/kropath/kropath-controller/issues/54)) ([ae8e9a1](https://github.com/kropath/kropath-controller/commit/ae8e9a1111bc582c586e652d8cbb2ce1b6898f18))

## [0.2.0](https://github.com/kropath/kropath-controller/compare/v0.1.0...v0.2.0) (2026-08-17)


### Features

* **KRO-555:** add APIGatewayConfig effectiveConfig cascade ([#49](https://github.com/kropath/kropath-controller/issues/49)) ([3b33d16](https://github.com/kropath/kropath-controller/commit/3b33d166a3075290d07859f86e53803fee4ae027))

## [0.1.0](https://github.com/kropath/kropath-controller/compare/v0.0.1...v0.1.0) (2026-08-14)


### Features

* **KRO-641:** add pr-title lint, OCI labels, and release process docs ([#46](https://github.com/kropath/kropath-controller/issues/46)) ([02fd4a0](https://github.com/kropath/kropath-controller/commit/02fd4a054ac8cd2ac67cc92cd7a16deb462a85e8))
* **KRO-656:** implement controller gaps 1-3 — dynamic global namespace, local default KPC, S3 parity ([#48](https://github.com/kropath/kropath-controller/issues/48)) ([1bbb71f](https://github.com/kropath/kropath-controller/commit/1bbb71fd4b49a09209f5da4f690968334689656c))
