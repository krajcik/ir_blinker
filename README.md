#Simple transparent overlay for Iracing
##which can blink on ABS activation and gear shifting

By default, the RPM values are taken from the SDK, but it is possible to enter your data into the vehicles.ini file.

## Compile

```bash
go build -o dashboard.exe main.go
```

## Run

Default host and port:
```bash
dashboard.exe
```

With custom host and port:
```bash
dashboard.exe -addr=192.168.1.100:8888
```
