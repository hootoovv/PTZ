import PTZControl
from config import loadConfig
import time

if __name__ == '__main__':
  ip, port, username, password = loadConfig()
  ptz = PTZControl.PTZControl(ip, port, username, password)

  configs = ptz.get_configs()
  print(configs)

  streams = []
  for config in configs['data']['Streams']:
    streams.append(config['Name'])

  print(streams)
  ptz.set_profile(streams[0])
  print("Set profile to: " + streams[0] + ".")

  uri = ptz.get_stream_uri()
  print(uri)

  ptz.set_profile(streams[1])
  print("Set profile to: " + streams[1] + ".")

  uri = ptz.get_stream_uri()
  print(uri)


  pos = ptz.get_position()
  print(pos)

  ptz.goto_home()
  print("Set PTZ to home position")
  time.sleep(1)
  
  move = ptz.is_moving()
  print(move)
  time.sleep(10)

  presets = ptz.get_presets()
  print(presets)

  ptz.goto_preset("2")
  print("Set PTZ to preset 2")
  time.sleep(10)

  ptz.goto_preset("1")
  print("Set PTZ to preset 1")
  time.sleep(10)

  ptz.goto_home()
  print("Set PTZ to home position")
  time.sleep(1)
  
  move = ptz.is_moving()
  print(move)
  # time.sleep(1)

  ptz.ptz_stop()
  print("Stop PTZ")
  time.sleep(0.5)

  move = ptz.is_moving()
  print(move)
  time.sleep(5)

  ptz.goto_home()
  print("Set PTZ to home position")
  time.sleep(10)


  ptz.goto_position(-0.5, 0.5, 0.5, 0.5, 0.5, 0.5)
  print("Set PTZ to p=-0.5, t=0.5, z=0.5, p_speed=0.5, t_speed=0.5, z_speed=0.5")
  time.sleep(2)

  move = ptz.is_moving()
  print(move)
  time.sleep(2)

  move = ptz.is_moving()
  print(move)

  ptz.move_relative_position(0.1, 0.1)
  print("Set PTZ to p=0.1, t=0.1")
  time.sleep(10)

  ptz.move_relative_position(0.1, 0.1)
  print("Set PTZ to p=0.1, t=0.1")
  time.sleep(10)

  ptz.move_relative_position(-0.4, -0.4)
  print("Set PTZ to p=-0.4, t=-0.4")
  time.sleep(10)

  ptz.goto_preset("2")
  print("Set PTZ to preset 2")

