# IP Camera PTZ Control (ONVIF) - Go

Keywords: Go, Gin, Javascript, ONVIF, REST API, Vue3, Element Plus, JQuery

So far only support H264 and H265 decode. It depends libavcodec, so need cgo support. Tested on WSL 1.0 env

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

## Files

* ptzcontrol.go - PTZ Control for ONVIF IP Cameras (tested on TPLink cam)

* server.go - Rest API server with session control, HTTP client can call snapshot API to get base64 encoded jpeg image

* server_gin.go - Gin version Rest API server

* session.go - session control, streaming rtsp video

* main.go - main program with ptzcontrol test

* h264_decoder.go - H264 decoder (wrapped by libavcodec, need cgo support)

* h265_decoder.go - H265 decoder (wrapped by libavcodec, need cgo support)

* static/index.html - web page for PTZ control and preview

## Todo

need catching panic in session.go.

## Notes

Tested on TPLink IP Camera.