# nyaa-cli
Terminal user interface for nyaa.si with support of [peerflix](https://github.com/mafintosh/peerflix).

Peerflix can be enabled with the `--peerflix` flag. By default the tool will only download the torrent file to the current directory (or if specified by `--dir`).

https://user-images.githubusercontent.com/7271496/164914020-5dd0c1cd-30c2-4bac-b539-5a0ccfc701ef.mp4

# Requirements
If you want the peerflix feature, you will need to install [peerflix](https://github.com/mafintosh/peerflix) (`npm i -g peerflix`).

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
