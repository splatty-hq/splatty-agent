# Changelog

## [1.2.0](https://github.com/splatty-hq/splatty-agent/compare/v1.1.0...v1.2.0) (2026-10-08)


### Features

* collect disk I/O and network rates on macOS ([d4ea83a](https://github.com/splatty-hq/splatty-agent/commit/d4ea83a514a2d6a313c794e62e611fc39a6e07f8))
* default SPLATTY_URL to https://splatty.app ([4baf8cc](https://github.com/splatty-hq/splatty-agent/commit/4baf8cce33bb8b91cbfd2a9a541870763a080273))
* report disk I/O and network packet rates ([4ad6d6c](https://github.com/splatty-hq/splatty-agent/commit/4ad6d6c9f86618fc945dd1493591867af7f9c833))
* stamp builds with the release tag and add --version ([7677b92](https://github.com/splatty-hq/splatty-agent/commit/7677b925cc4c53a801c316b7ef287c074f4f172f))


### Bug Fixes

* correct module path so the agent is go-installable ([f5c1b1f](https://github.com/splatty-hq/splatty-agent/commit/f5c1b1f20dbdb299e6417aa69c2b4e32d40b3176))
* **transport:** report the real build version in the User-Agent ([590354d](https://github.com/splatty-hq/splatty-agent/commit/590354d78e5eda431e2d9aeb25bd8fcfa3ed104c))

## [1.1.0](https://github.com/splatty-hq/splatty-agent/compare/v1.0.0...v1.1.0) (2026-08-05)


### Features

* docker container metrics, rename SPLATTY_DSN to SPLATTY_SERVER_TOKEN ([bd765ea](https://github.com/splatty-hq/splatty-agent/commit/bd765ea1824fd37d4507a6591187d2a88dc80d74))
* emit swap.used_bytes on linux ([495ed36](https://github.com/splatty-hq/splatty-agent/commit/495ed366521b5b266c7e31bfc5c86fce675058e4))

## 1.0.0 (2026-06-23)


### Features

* darwin cpu and used/available memory via top + vm_stat ([e570a6d](https://github.com/splatty-hq/splatty-agent/commit/e570a6d961c5527f77bd6151e06fa9f82ad7c768))
* initial splatty-agent (extracted from splatty repo) ([cf1abf5](https://github.com/splatty-hq/splatty-agent/commit/cf1abf55b75e74e0c89534aaf253bad7f58dc647))
* native macos collectors via sysctl (load, mem total, swap, disk) ([9d20fe1](https://github.com/splatty-hq/splatty-agent/commit/9d20fe1ee187bf933626dd6fb15128ab5b324dac))
* release workflow and install.sh with optional systemd setup ([0440936](https://github.com/splatty-hq/splatty-agent/commit/04409361f4e2cb03d804619d2ff286f88fde0603))


### Bug Fixes

* dedup collect errors so missing /proc on darwin doesn't spam every cycle ([fa832d8](https://github.com/splatty-hq/splatty-agent/commit/fa832d8cf7c2cdd3751d53ebd4f67303492b98bf))


### Performance Improvements

* explicit http transport keep-alive ([9562aca](https://github.com/splatty-hq/splatty-agent/commit/9562aca809f33f0fd3068e6c7d5bc8f05c715913))
* gzip outbound metric batches ([1b64c96](https://github.com/splatty-hq/splatty-agent/commit/1b64c965d5bd7619e73594bca683084639bba146))
