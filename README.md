# kcptun-walk

![Cat](./assets/icon.ico)

Windows only, for now. [Icon from here.](https://www.iconfinder.com/icons/182507/cat_walk_icon)

## Build

### Build executable
```console
go build -ldflags -H=windowsgui .
```

### Build with version
```console
go build -ldflags="-X main.tagName=$(git describe --tags --abbrev=0) -X main.sha=$(git rev-parse --short HEAD)" -H=windowsgui
```

## Set exe icon
With [rsrc](https://github.com/akavel/rsrc):
```console
rsrc.exe -ico assets/icon.ico
```
