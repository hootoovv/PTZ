import PTZControl
import cv2
import time
import numpy as np
from config import loadConfig

if __name__ == '__main__':
  ip, port, username, password = loadConfig()
  ptz = PTZControl.PTZControl(ip, port, username, password)

  configs = ptz.get_configs()
  print(configs)

  streams = []
  sizes = []
  for config in configs['data']['Streams']:
    streams.append(config['Name'])
    sizes.append((config['Video']['Resolution']['Width'], config['Video']['Resolution']['Height']))

  print(streams)
  print(sizes)
  win_size_list = [(1280, 720), (640, 480)]


  STREAM_ID = 1 # 0
  stream = streams[STREAM_ID]
  ptz.set_profile(stream)
  print("Set profile to: " + stream + ".")

  uri = ptz.get_stream_uri()
  print(uri)

  pos = ptz.get_position()
  print(pos)

  rtsp = uri['data']['Uri'].replace("rtsp://", "rtsp://"+ptz.username+":"+ptz.password+"@")
  print(rtsp)

  # 读取视频
  cap = cv2.VideoCapture(rtsp)

  cv2.namedWindow(stream, 0)
  cv2.resizeWindow(stream, win_size_list[STREAM_ID])

  # used to record the time when we processed last frame 
  prev_frame_time = 0  
  # used to record the time at which we processed current frame 
  new_frame_time = 0
  # font which we will be using to display FPS 
  font = cv2.FONT_HERSHEY_SIMPLEX 

  frame_drop_num=3
  frame_drop_counter =0
  frame_drop_time=[0,0,0]

  while True:
    # time when we finish processing for this frame 
    new_frame_time = time.time()
    # Calculating the fps
    # fps will be number of frame processed in given time frame 
    # since their will be most of time error of 0.001 second 
    # we will be subtracting it to get more accurate result 
    fps = 1/(new_frame_time-prev_frame_time)
    
    frame_drop_time.append(fps)
    frame_drop_time.pop(0)      
    prev_frame_time = new_frame_time
    
    # 读取一帧视频
    ret, frame = cap.read()
    if frame is None:
        break
    
    # cv2.putText(frame, "Frame No.: %d" % cap.get(cv2.CAP_PROP_POS_FRAMES), (20, 80), font, 1, (100, 255, 0), 2, cv2.LINE_AA)
    
    # converting the fps into integer 
    fps = np.mean(frame_drop_time)
    fps = str('{:.1f}'.format(fps) )  
    # converting the fps to string so that we can display it on frame 
    # by using putText function 
    fps = str('fps:'+ fps)
    cv2.putText(frame, fps, (20, 40), font, 1, (100, 255, 0), 2, cv2.LINE_AA)

    cv2.imshow(streams[STREAM_ID], frame)

    # if cv2.waitKey(1) == ord('q'):
    #   break
    key = cv2.waitKeyEx(1)

    if key == ord("q") or key == 27:  # exit
        break
    elif key == ord("u") or key == 2490368: # up
      ptz.move_relative_position(0,-0.1)
    elif key == ord("d") or key == 2621440: # down
      ptz.move_relative_position(0,0.1)
    elif key == ord("l") or key == 2424832: # left
      ptz.move_relative_position(-0.1,0)
    elif key == ord("r") or key == 2555904: # right
      ptz.move_relative_position(0.1,0)
    elif key == ord("h") or key == 2359296: # home
      ptz.goto_home()
    elif key == ord("1"): # preset 1
      ptz.goto_preset(1)
    elif key == ord("2"): # preset 2
      ptz.goto_preset(2)
    elif key == ord("s"): # stop
      ptz.ptz_stop()



  # 完成所有操作后，释放捕获器
  cap.release()

  cv2.destroyAllWindows()

 