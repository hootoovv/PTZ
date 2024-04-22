# IP Camera PTZ Control (ONVIF) - Go

Keywords: Go, Javascript, ONVIF, REST API, Vue3, Element Plus, JQuery

## web server for PTZ control/preview

1. change config.yaml to your web cam parameters

2. start web server

```shell
sudo apt install pkg-config
sudo apt install libavcodec-dev
sudo apt install libavutil-dev
sudo apt install libswscale-dev
sudo ldconfig

go build .
./ptz_go
```

3. start web browser to connect to http://localhost:8000

## Todo

need catching panic in session.go.

## Notes

Tested on TPLink IP Camera.