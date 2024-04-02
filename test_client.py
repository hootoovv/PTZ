import requests
import requests_unixsocket
import platform
import time
from config import loadConfig



sys = platform.system()
if sys == "Linux":
  url = 'http+unix://%2Ftmp%2Ffastapi.sock'
else:
  url = "http://127.0.0.1:8000"

requests_unixsocket.monkeypatch()

def test_root():
  response = requests.get(url+'/ptz')
  print(response.json())  

def test_connect(route):
  ip, port, username, password = loadConfig()
  payload = {"ip": ip, "port": port, "username": username, "password": password}
  response = requests.post(url + route, json = payload)

  json = response.json()

  if json["code"] == 200:
    # print(json["data"]["session_id"])
    cookie = response.cookies
  else:
    print("Error: " + str(json["code"]) + " - " + json["message"])
    return ''

  return cookie

def test_snapshot(route, cookie):
  response = requests.get(url + route, cookies = cookie)

  json = response.json()

  if json["code"] == 200:
    print("Got snapshot, size: " + str(int(json["data"]["w"])) + "x" + str(int(json["data"]["h"])))

    base64 = json["data"]["image"]

    with open('base64.txt', "w") as f:
      f.write(base64)
  else:
    print("Error: " + str(json["code"]) + " - " + json["message"])
    return ''

def test_get(route, cookie):
  response = requests.get(url + route, cookies = cookie)
  json = response.json()
  print(json)

def test_post(route, cookie, payload=None):
  if payload == None:
    response = requests.post(url + route, cookies = cookie)
  else:
    response = requests.post(url + route, cookies = cookie, json = payload)
  json = response.json()
  print(json)

if __name__ == '__main__':
  print("Test Home")
  test_root()
  print("OK\n")

  print("Test Connect")
  cookie = test_connect("/ptz/connect")
  print("OK\n")

  print("Test Snapshot")
  test_snapshot("/snapshot", cookie)

  time.sleep(5)

  test_snapshot("/snapshot", cookie)

  print("Test Get Configs")
  test_get("/ptz/config", cookie)

  print("Test Get Presets")
  test_get("/ptz/presets", cookie)
  
  print("Test Get Position")
  test_get("/ptz/position", cookie)

  print("Test Post profile")
  payload = {"profile": "mainStream"}
  # payload = {"profile": "minorStream"}
  test_post("/ptz/profile", cookie, payload)

  time.sleep(1)

  test_snapshot("/snapshot", cookie)

  time.sleep(5)
  test_snapshot("/snapshot", cookie)

  time.sleep(5)
  test_snapshot("/snapshot", cookie)

  # payload = {"profile": "mainStream"}
  payload = {"profile": "minorStream"}
  test_post("/ptz/profile", cookie, payload)

  time.sleep(1)
  test_snapshot("/snapshot", cookie)

  time.sleep(5)
  test_snapshot("/snapshot", cookie)

  time.sleep(5)
  test_snapshot("/snapshot", cookie)

  payload = {"preset": 1}
  test_post("/ptz/goto/preset", cookie, payload)

  time.sleep(1)
  test_get("/ptz/moving", cookie)

  time.sleep(1)
  test_post("/ptz/stop", cookie)

  time.sleep(0.5)
  test_get("/ptz/position", cookie)

  test_post('/ptz/goto/home', cookie)
  time.sleep(5)

  payload = {"Pan": 0.1, "Tilt": 0.1, "Zoom": 0, "PanSpeed": 1, "TiltSpeed": 1, "ZoomSpeed": 1}
  test_post('/ptz/goto/position', cookie, payload)
  time.sleep(10)

  payload = {"Pan": -0.1, "Tilt": -0.1, "Zoom": 0, "PanSpeed": 1, "TiltSpeed": 1, "ZoomSpeed": 1}
  test_post('/ptz/move/relative', cookie, payload)
  time.sleep(10)

  time.sleep(5)
  payload = {"preset": 2}
  test_post("/ptz/goto/preset", cookie, payload)


