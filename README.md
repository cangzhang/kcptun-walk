# kcptun-walk

Windows only, for now. [Icon from here.](https://www.iconfinder.com/icons/182507/cat_walk_icon)

## Build
```console
go build -ldflags -H=windowsgui .
```

## Set exe icon
With [rsrc](https://github.com/akavel/rsrc):
```console
rsrc.exe -ico assets/icon.ico
```
