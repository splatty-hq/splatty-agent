# Changelog

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
