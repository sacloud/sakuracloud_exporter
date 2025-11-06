# Changelog

## [0.19.2](https://github.com/sacloud/sakuracloud_exporter/compare/0.19.1...0.19.2) - 2025-11-06
- iaas-api-go v1.18.1 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/251

## [0.19.1](https://github.com/sacloud/sakuracloud_exporter/compare/0.19.0...0.19.1) - 2025-10-08
- Push Docker image with both latest and release tag by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/239

## [0.19.0](https://github.com/sacloud/sakuracloud_exporter/compare/0.18.6...0.19.0) - 2025-10-08
- go: bump github.com/stretchr/testify from 1.9.0 to 1.10.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/210
- deprecated funcの対応 by @hekki in https://github.com/sacloud/sakuracloud_exporter/pull/214
- ProxyLBCollectorのcert_info, cert_expireのテストを実装 by @hekki in https://github.com/sacloud/sakuracloud_exporter/pull/215
- go: bump golang.org/x/crypto from 0.17.0 to 0.31.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/212
- go: bump github.com/sacloud/webaccel-api-go from 1.1.6 to 1.2.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/211
- go: bump github.com/sacloud/iaas-api-go from 1.11.2 to 1.14.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/216
- go: bump github.com/prometheus/client_golang from 1.19.0 to 1.20.5 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/207
- go: bump github.com/sacloud/iaas-service-go from 1.9.2 to 1.10.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/197
- Copyright 2025 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/218
- ci: introduce tagpr and align workflows with other repos by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/223
- VPCルータにおいて任意の位置にNICを作成する場合を考慮する by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/222
- go: bump github.com/sacloud/packages-go from 0.0.10 to 0.0.11 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/217
- Use iaas.SakuraCloudZones by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/232
- golangci-lint v2 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/233
- sacloud/iaas-service-go v1.14.1 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/234
- ci: bump actions/setup-go from 5 to 6 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/225
- go: bump github.com/cloudflare/circl from 1.3.7 to 1.6.1 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/235
- go: bump github.com/prometheus/client_model from 0.6.1 to 0.6.2 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/228
- go: bump github.com/alexflint/go-arg from 1.5.1 to 1.6.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/227
- go: bump github.com/prometheus/client_golang from 1.20.5 to 1.23.2 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/226
- docker: golang:1.25 & alpine:3.22 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/236
- publish docker image with tagpr by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/238
- go: bump github.com/sacloud/webaccel-api-go from 1.2.0 to 1.3.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/237

## [0.18.6](https://github.com/sacloud/sakuracloud_exporter/compare/0.18.5...0.18.6) - 2024-11-28
- go: bump google.golang.org/protobuf from 1.32.0 to 1.33.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/195
- go: bump github.com/prometheus/client_model from 0.6.0 to 0.6.1 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/196
- ci: bump docker/build-push-action from 5 to 6 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/208
- go: bump github.com/hashicorp/go-retryablehttp from 0.7.5 to 0.7.7 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/205
- ci: bump goreleaser/goreleaser-action from 5 to 6 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/200
- GoReleaser v2 by @hekki in https://github.com/sacloud/sakuracloud_exporter/pull/209
- go: bump github.com/alexflint/go-arg from 1.4.3 to 1.5.1 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/206

## [0.18.5](https://github.com/sacloud/sakuracloud_exporter/compare/0.18.4...0.18.5) - 2024-10-28
- update dependencies by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/194
- go: bump github.com/stretchr/testify from 1.8.4 to 1.9.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/193
- go: bump github.com/prometheus/client_golang from 1.17.0 to 1.19.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/192
- go: bump github.com/prometheus/client_model from 0.5.0 to 0.6.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/191
- go: bump github.com/cloudflare/circl from 1.3.3 to 1.3.7 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/190
- Update README - added note on bill and coupon by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/201
- VSCodeのconfigをignore by @hekki in https://github.com/sacloud/sakuracloud_exporter/pull/203
- billing APIをキャッシュする by @hekki in https://github.com/sacloud/sakuracloud_exporter/pull/202
- coupon APIをキャッシュする by @hekki in https://github.com/sacloud/sakuracloud_exporter/pull/204

## [0.18.4](https://github.com/sacloud/sakuracloud_exporter/compare/0.18.3...0.18.4) - 2023-12-12
- go: bump github.com/prometheus/client_model from 0.3.0 to 0.4.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/155
- go: bump github.com/sacloud/iaas-service-go from 1.7.0 to 1.8.2 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/154
- go: bump github.com/prometheus/client_golang from 1.14.0 to 1.15.1 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/156
- go: bump github.com/sacloud/api-client-go from 0.2.7 to 0.2.8 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/159
- go: bump github.com/sacloud/packages-go from 0.0.8 to 0.0.9 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/158
- go: bump github.com/stretchr/testify from 1.8.2 to 1.8.4 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/162
- go: bump github.com/sacloud/iaas-api-go from 1.10.0 to 1.11.1 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/163
- go: bump github.com/sacloud/iaas-service-go from 1.8.2 to 1.9.1 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/164
- go: bump github.com/prometheus/client_golang from 1.15.1 to 1.16.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/165
- go 1.21 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/169
- go-kit/log -> log/slog by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/168
- Update example-fake-store.json by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/178
- ci: bump docker/setup-buildx-action from 2 to 3 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/174
- ci: bump docker/build-push-action from 4 to 5 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/173
- ci: bump goreleaser/goreleaser-action from 4 to 5 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/172
- ci: bump crazy-max/ghaction-import-gpg from 5 to 6 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/171
- ci: bump actions/checkout from 3 to 4 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/170
- go: bump github.com/prometheus/client_model from 0.4.0 to 0.5.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/176
- go: bump github.com/prometheus/client_golang from 1.16.0 to 1.17.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/175
- go: bump github.com/sacloud/api-client-go from 0.2.8 to 0.2.9 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/166
- ci: bump docker/login-action from 2 to 3 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/179
- ci: bump docker/metadata-action from 4 to 5 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/180
- ci: bump docker/setup-qemu-action from 2 to 3 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/181
- ci: bump actions/setup-go from 4 to 5 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/182
- go: bump github.com/sacloud/iaas-service-go from 1.9.1 to 1.9.2 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/187
- go: bump github.com/sacloud/webaccel-api-go from 1.1.5 to 1.1.6 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/183

## [0.18.3](https://github.com/sacloud/sakuracloud_exporter/compare/0.18.2...0.18.3) - 2023-03-30
- go: bump github.com/sacloud/iaas-service-go from 1.6.1 to 1.7.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/147
- go: bump github.com/sacloud/webaccel-api-go from 1.1.4 to 1.1.5 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/148
- go: bump github.com/sacloud/iaas-api-go from 1.9.0 to 1.9.1 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/149

## [0.18.2](https://github.com/sacloud/sakuracloud_exporter/compare/0.18.1...0.18.2) - 2023-03-28
- 利用していないワークフローを除去 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/105
- go 1.19 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/107
- sacloud/makefile v0.0.7 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/108
- go: bump github.com/sacloud/packages-go from 0.0.5 to 0.0.6 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/110
- go: bump github.com/sacloud/iaas-service-go from 1.3.1 to 1.3.2 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/111
- go: bump github.com/sacloud/webaccel-api-go from 1.1.3 to 1.1.4 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/112
- go: bump github.com/sacloud/iaas-api-go from 1.4.1 to 1.5.1 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/116
- go: bump github.com/prometheus/client_model from 0.2.0 to 0.3.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/117
- go: bump github.com/stretchr/testify from 1.8.0 to 1.8.1 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/120
- go: bump github.com/sacloud/iaas-service-go from 1.3.2 to 1.4.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/118
- go: bump github.com/prometheus/client_golang from 1.13.0 to 1.14.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/122
- go: bump github.com/sacloud/api-client-go from 0.2.3 to 0.2.4 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/123
- go: bump github.com/sacloud/iaas-api-go from 1.6.0 to 1.6.1 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/124
- go: bump github.com/sacloud/packages-go from 0.0.6 to 0.0.7 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/126
- ci: bump goreleaser/goreleaser-action from 3 to 4 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/127
- go: bump github.com/sacloud/iaas-service-go from 1.4.0 to 1.5.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/129
- go: bump github.com/sacloud/iaas-api-go from 1.7.0 to 1.7.1 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/130
- copyright 2023 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/131
- go: bump github.com/sacloud/iaas-service-go from 1.5.0 to 1.6.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/133
- go: bump github.com/sacloud/iaas-api-go from 1.8.0 to 1.8.1 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/134
- go: bump github.com/sacloud/iaas-api-go from 1.8.1 to 1.8.3 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/136
- ci: bump docker/build-push-action from 3 to 4 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/137
- go: bump github.com/stretchr/testify from 1.8.1 to 1.8.2 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/139
- go: bump golang.org/x/crypto from 0.0.0-20220214200702-86341886e292 to 0.1.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/140
- go: bump github.com/sacloud/iaas-service-go from 1.6.0 to 1.6.1 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/141
- go 1.20 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/142
- ci: bump actions/setup-go from 3 to 4 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/143
- go: bump github.com/sacloud/iaas-api-go from 1.8.3 to 1.9.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/144

## [0.18.1](https://github.com/sacloud/sakuracloud_exporter/compare/0.18.0...0.18.1) - 2022-09-05
- go 1.18 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/101

## [0.18.0](https://github.com/sacloud/sakuracloud_exporter/compare/0.17.1...0.18.0) - 2022-09-05
- sacloud/go-template@v0.0.2 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/65
- Dockerfile更新 - make buildの出力先変更対応 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/72
- ci: bump docker/login-action from 1 to 2 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/66
- go: bump github.com/sacloud/iaas-service-go from 1.0.0 to 1.1.3 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/67
- go: bump github.com/sacloud/api-client-go from 0.1.0 to 0.2.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/70
- go: bump github.com/go-kit/kit from 0.8.0 to 0.12.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/71
- go: bump github.com/stretchr/testify from 1.7.1 to 1.8.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/69
- sacloud/go-template@v0.0.5 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/79
- go: bump github.com/sacloud/iaas-api-go from 1.1.2 to 1.2.1 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/77
- go: bump github.com/sacloud/iaas-service-go from 1.1.3 to 1.2.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/76
- go: bump github.com/alexflint/go-arg from 1.0.0 to 1.4.3 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/73
- go: bump github.com/prometheus/client_golang from 1.11.0 to 1.12.2 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/74
- Configのテスト by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/81
- e2eテスト by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/83
- crazy-max/ghaction-import-gpg@v5でのパラメータ名変更対応 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/89
- 専有ホストのIDをラベルに追加 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/90
- go: bump github.com/prometheus/client_golang from 1.12.2 to 1.13.0 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/91
- go: bump github.com/sacloud/api-client-go from 0.2.0 to 0.2.1 by @dependabot[bot] in https://github.com/sacloud/sakuracloud_exporter/pull/92
- iaas-api-go@v1.3.1 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/95
- APIクライアントのパラメータでデフォルト値を利用 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/96
- ウェブアクセラレータ対応 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/97
- 各リソースの値を参照する前にAvailabilityを確認する by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/98
- 当月の請求金額: sakuracloud_bill_amount by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/99
- アプライアンスでのメンテナンス情報 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/100

## [0.17.1](https://github.com/sacloud/sakuracloud_exporter/compare/0.17.0...0.17.1) - 2022-05-17
- libsacloud/v2からiaas-api-goへ切り替え by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/64

## [0.17.0](https://github.com/sacloud/sakuracloud_exporter/compare/0.16.0...0.17.0) - 2022-01-11
- Annual update of the copyright notice by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/60
- メンテナンス情報のみ取得できるオプションを追加 by @pitan in https://github.com/sacloud/sakuracloud_exporter/pull/61
- Enable ChangeLog on release by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/62

## [0.16.0](https://github.com/sacloud/sakuracloud_exporter/compare/0.15.1...0.16.0) - 2021-12-07
- upgrade dependencies by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/57
- VPCRouter: sakuracloud_vpc_router_cpu_time by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/58
- Go 1.17 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/59

## [0.15.1](https://github.com/sacloud/sakuracloud_exporter/compare/0.15.0...0.15.1) - 2021-10-25
- libsacloud v2.27.1 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/55

## [0.15.0](https://github.com/sacloud/sakuracloud_exporter/compare/0.14.0...0.15.0) - 2021-07-30
- VPCRouter: sakuracloud_vpc_router_session_analysis by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/54

## [0.14.0](https://github.com/sacloud/sakuracloud_exporter/compare/0.13.2...0.14.0) - 2021-03-25
- Go 1.16 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/47
- libsacloud v2.14.1 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/48
- switch from ghr to goreleaser by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/50
- docker/build-push-action@v2 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/51
- fix some build settings by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/52

## [0.13.2](https://github.com/sacloud/sakuracloud_exporter/compare/0.13.1...0.13.2) - 2021-01-08
- Some updates by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/46

## [0.13.1](https://github.com/sacloud/sakuracloud_exporter/compare/0.13.0...0.13.1) - 2020-10-02
- Add workflow file for publishing to GitHub Container Registry by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/42
- Add workflow file for publishing latest image to ghcr by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/43
- Update README.md - Use GitHub Container Registry instead of DockerHub by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/44

## [0.13.0](https://github.com/sacloud/sakuracloud_exporter/compare/0.12.0...0.13.0) - 2020-10-01
- ESME by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/41

## [0.12.0](https://github.com/sacloud/sakuracloud_exporter/compare/0.11.1...0.12.0) - 2020-08-20
- tk1b zone by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/39

## [0.11.1](https://github.com/sacloud/sakuracloud_exporter/compare/0.11.0...0.11.1) - 2020-03-11
- libsacloud v2.1.9 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/37
- Go 1.14 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/38

## [0.11.0](https://github.com/sacloud/sakuracloud_exporter/compare/0.10.0...0.11.0) - 2020-01-30
- Libsacloud v2.0.0 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/34
- Introduce golangci-lint by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/35
- Introduce GitHub Actions by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/36

## [0.10.0](https://github.com/sacloud/sakuracloud_exporter/compare/0.9.0...0.10.0) - 2020-01-22
- Add systemd examples by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/32
- Supports LocalRouter API by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/33

## [0.9.0](https://github.com/sacloud/sakuracloud_exporter/compare/0.8.0...0.9.0) - 2020-01-16
- Use libsacloud v2.0.0-rc4 by @yamamoto-febc in https://github.com/sacloud/sakuracloud_exporter/pull/31
