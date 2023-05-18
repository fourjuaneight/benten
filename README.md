# Benten

A simple utility to backup media to a B2 instance.

For help on how to use the app, you can run `benten -h`.

## Usage
There are 3 ways to run the script:

### [Gorun](https://github.com/erning/gorun#how-to-build-and-install-gorun-from-source)
```sh
make run
# script should run from root of repo
./benten.go
```

### Local Binary
```sh
make build
# binary should be accessible from the root of the repo
./benten
```

### GOPATH Binary
```sh
make install
# binary should be accessible from anywhere
benten
```
