from fastapi import FastAPI, Response, Cookie
from fastapi.responses import FileResponse
from fastapi.staticfiles import StaticFiles
# from fastapi.middleware.cors import CORSMiddleware
import uvicorn
import platform
import PTZSession
import time
import threading

sessions = dict()
app = FastAPI()

app.mount("/js", StaticFiles(directory="static/js"), name="js")
app.mount("/css", StaticFiles(directory="static/css"), name="css")

# origins = [
#     "http://localhost",
#     "http://localhost:8000",
# ]

# app.add_middleware(
#     CORSMiddleware,
#     allow_origins=origins,
#     allow_credentials=True,
#     allow_methods=["*"],
#     allow_headers=["*"],
# )

# session_id: str = Cookie(None)

@app.get("/")
async def home():
  return FileResponse("static/index.html")

@app.get("/favicon.ico")
async def home():
  return FileResponse("static/favicon.ico")

@app.get("/ptz")
async def root():
  return {"message": "PTZ Server"}

@app.get("/ptz/config")
async def ptz_get_config(session_id: str = Cookie(None)):
  global sessions

  try:
    session = sessions[session_id]
    session.activate_session()
    return session.ptz.get_configs()
  except:
    return {"code": 401, "message": "Unauthorized", "data": {}}
  
@app.get("/ptz/presets")
async def ptz_get_presets(session_id: str = Cookie(None)):
  global sessions

  try:
    session = sessions[session_id]
    session.activate_session()
    return session.ptz.get_presets()
  except:
    return {"code": 401, "message": "Unauthorized", "data": {}}
@app.get("/ptz/position")
async def ptz_get_position(session_id: str = Cookie(None)):
  global sessions

  try:
    session = sessions[session_id]
    session.activate_session()
    return session.ptz.get_position()
  except:
    return {"code": 401, "message": "Unauthorized", "data": {}}

@app.get("/snapshot")
async def get_snapshot(session_id: str = Cookie(None)):
  global sessions

  try:
    session = sessions[session_id]
    session.activate_session()
    return session.snapshot()
  except:
    return {"code": 401, "message": "Unauthorized", "data": {}}

@app.post("/ptz/connect")
async def ptz_start(data: dict, response: Response):
  global sessions

  # print(data)

  sid = ''

  for id in sessions.keys():
    session = sessions[id]
    if data['ip'] == session.ip and data['port'] == session.port:
      session.activate_session()
      # Session exists, return session_id
      print("Session " + session.ip + ":" + str(session.port) + " exists.")
      sid = session.session_id
      break

  if sid == '':
    session = PTZSession.PTZSession()
    session.create_session(data['ip'], data['port'], data['username'], data['password'])
    sid = session.session_id
    if sid == '':
      return {'code': 404, 'message': 'Session creation failed', 'data': {}}
    else:
      sessions[sid] = session

  response.set_cookie(key="session_id", value=sid)

  return {'code': 200, 'message': 'Session started', 'data': {'session_id': sid}}

@app.post("/ptz/profile")
async def ptz_set_profile(data: dict, session_id: str = Cookie(None)):
  global sessions

  try:
    session = sessions[session_id]
    session.activate_session()
    profile = data["profile"]
    if profile != session.ptz.profile:
      session.change_profile(profile)
    return session.ptz.set_profile(data['profile'])
  except:
    return {"code": 401, "message": "Unauthorized", "data": {}}


@app.post("/ptz/move/relative")
async def move_relative_position(data: dict, session_id: str = Cookie(None)):
  global sessions

  try:
    session = sessions[session_id]
    session.activate_session()
    p, t, z = data['Pan'], data['Tilt'], data['Zoom']
    ps, ts, zs = data['PanSpeed'], data['TiltSpeed'], data['ZoomSpeed']
    return session.ptz.move_relative_position(p, t, z, ts, ps, zs)
  except:
    return {"code": 401, "message": "Unauthorized", "data": {}}

@app.post("/ptz/goto/position")
async def goto_position(data: dict, session_id: str = Cookie(None)):
  global sessions

  try:
    session = sessions[session_id]
    session.activate_session()
    p, t, z = data['Pan'], data['Tilt'], data['Zoom']
    ps, ts, zs = data['PanSpeed'], data['TiltSpeed'], data['ZoomSpeed']
    return session.ptz.goto_position(t, p, z, ts, ps, zs)
  except:
    return {"code": 401, "message": "Unauthorized", "data": {}}

@app.post("/ptz/goto/home")
async def goto_home(session_id: str = Cookie(None)):
  global sessions

  try:
    session = sessions[session_id]
    session.activate_session()
    return session.ptz.goto_home()
  except:
    return {"code": 401, "message": "Unauthorized", "data": {}}

@app.post("/ptz/goto/preset")
async def goto_preset(data: dict, session_id: str = Cookie(None)):
  global sessions

  try:
    session = sessions[session_id]
    session.activate_session()
    return session.ptz.goto_preset(data['preset'])
  except:
    return {"code": 401, "message": "Unauthorized", "data": {}}

@app.post("/ptz/stop")
async def relative_move(session_id: str = Cookie(None)):
  global sessions

  try:
    session = sessions[session_id]
    session.activate_session()
    return session.ptz.ptz_stop()
  except:
    return {"code": 401, "message": "Unauthorized", "data": {}}

@app.get("/ptz/moving")
async def is_moving(session_id: str = Cookie(None)):
  global sessions

  try:
    session = sessions[session_id]
    session.activate_session()
    return session.ptz.is_moving()
  except:
    return {"code": 401, "message": "Unauthorized", "data": {}}
  
def check_session_thread_func():
  global sessions

  while True:
    for id in sessions.keys():
      session = sessions[id]
      if session.is_session_end():
        sessions.pop(id)
        break
    
    time.sleep(1)

# uvicorn server:app --reload
# uvicorn server:app --uds /tmp/fastapi.sock

if __name__ == '__main__':

  check_session_thread = threading.Thread(target=check_session_thread_func)
  check_session_thread.start()

  sys = platform.system()
  if sys == "Linux":
    uvicorn.run(app=app, uds="/tmp/ptz_fastapi.sock")
  else:
    uvicorn.run(app=app, host="127.0.0.1", port=8000)