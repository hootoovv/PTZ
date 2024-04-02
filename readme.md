# IP Camera PTZ Control (ONVIF)

Keywords: Python, Javascript, ONVIF, FASTAPI, OpenCV, REST API, Vue3, Element Plus, JQuery

## web server for PTZ control/preview

1. change config.yaml to your web cam parameters

2. start web server

```shell

python rest_server.py
```

3. start web browser to connect to http://localhost:8000

## Files

* PTZControl.py - PTZ Control for ONVIF IP Cameras (tested on TPLink cam)

* rest_server.py - Rest API server with session control, HTTP client can call snapshot API to get base64 encoded jpeg image

* test_control.py - test PTZ control class

* test_local.py - opencv window to monitor IP Camera's video stream(rtsp), keyboard to control PTZ: arrow keys etc. (detailed control key pls refer to code.)

* test_client.py - client to test Rest API

* static/index.html - web page for PTZ control and preview

## Notes

Tested on TPLink IP Camera.
