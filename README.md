# nyaa-cli

Terminal user interface for nyaa.si that directly runs the videos like it's streaming.

https://user-images.githubusercontent.com/7271496/164996729-0a2578d9-3b2d-4d2b-8e39-85388b811e98.mp4

# Requirements

The default video player is VLC. But you can specify another player by setting the `--player` option.

**Supported players**:
- vlc

# How to install

## From releases

Releases contains prebuilt binaries for Linux, macOS and Windows. You can download them at https://github.com/quantumsheep/nyaa-cli/releases.

## From GitHub Actions

Each commit is built and saved as GitHub Actions artifacts. You can download the latest version from the [GitHub Actions page](https://github.com/quantumsheep/nyaa-cli/actions?query=event%3Apush+is%3Asuccess+branch%3Amain).

## From sources

```bash
git clone https://github.com/quantumsheep/nyaa-cli.git
cd nyaa-cli
make
make install
```
