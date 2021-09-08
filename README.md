# kcptun-walk

![Cat](./assets/icon.ico)

Windows only, for now. [Icon from here.](https://www.iconfinder.com/icons/182507/cat_walk_icon)

## Run

1. rename your kcptun config as `config.json`
2. place it together with `kcptun-walk.exe` 
3. execute `kcptun-walk.exe`

## Build

### Build executable without cmd prompt 
```console
go build -ldflags -H=windowsgui
```

### Build with tag name/version
```console
go build -ldflags="-X main.tagName=$(git describe --tags --abbrev=0) -X main.sha=$(git rev-parse --short HEAD) -H=windowsgui"
```

## Set exe icon
With [rsrc](https://github.com/akavel/rsrc):
```console
rsrc.exe -ico assets/icon.ico
```
