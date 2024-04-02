import time
import threading
import PTZControl
import cv2
import uuid
import base64


class PTZSession:

  def __init__(self):
    self.session_id = str(uuid.uuid4())

    self.ptz = None

    self.video_thread = None
    self.video_thread_running = False
    self.session_thread = None

    self.frame = None
    self.frame_width = 0
    self.frame_height = 0

    self.timeout = 30
    self.last_time = 0
    self.session_end = False

  def create_session(self, ip, port, username, password):
    self.ip = ip
    self.port = port
    self.username = username
    self.password = password

    self.stop_video_thread()

    self.ptz = PTZControl.PTZControl(ip, port, username, password)
    if self.ptz.is_connected():
      res = self.ptz.get_stream_uri()
      
      rtsp_uri = res['data']['Uri'].replace("rtsp://", "rtsp://"+self.username+":"+self.password+"@")
      self.video_thread_running = True
      self.video_thread = threading.Thread(target=self.video_thread_func, args=(rtsp_uri,))
      self.video_thread.start()

      self.activate_session()

      self.session_thread = threading.Thread(target=self.session_thread_func)
      self.session_thread.start()

      return self.session_id
    else:
      return ''

  def change_profile(self, profile):
    # print("change profile")
    self.ptz.set_profile(profile)
    self.stop_video_thread()
    # print("cap thread exit")

    if self.ptz.is_connected():
      res = self.ptz.get_stream_uri()
      
      rtsp_uri = res['data']['Uri'].replace("rtsp://", "rtsp://"+self.username+":"+self.password+"@")
      # print(rtsp_uri)
      self.video_thread_running = True
      self.video_thread = threading.Thread(target=self.video_thread_func, args=(rtsp_uri,))
      self.video_thread.start()

      self.activate_session()

  def stop_video_thread(self):
    if self.video_thread is not None:
      self.video_thread_running = False
      self.video_thread.join()
      self.video_thread = None

  def video_thread_func(self, uri):
    cap = cv2.VideoCapture(uri)
    self.frame_width = cap.get(cv2.CAP_PROP_FRAME_WIDTH)
    self.frame_height = cap.get(cv2.CAP_PROP_FRAME_HEIGHT)

    # print(str(int(self.frame_width)) +  "x" + str(int(self.frame_height)))

    while self.video_thread_running:
      ret, frame = cap.read()
      if ret:
        self.frame = frame
      else:
        self.video_thread_running = False
        break

    cap.release()
    print("Video Stream Thread exit")

  def session_thread_func(self):
    print("Session Thread for " + self.ip + ":" + str(self.port) + " started")

    while (time.time() - self.last_time) < self.timeout:
      time.sleep(1)

    # Session timeout
    self.stop_video_thread()

    print("Session Thread for " + self.ip + ":" + str(self.port) + " end")

    self.session_end = True

  def is_session_end(self):
    return self.session_end
  
  def activate_session(self):
    self.last_time = time.time()

  def snapshot(self):
    if self.frame is not None:
      jpeg = cv2.imencode('.jpg', self.frame)[1]
      image_base64 = 'data:image/jpeg;base64,' + str(base64.b64encode(jpeg))[2:-1]
      return {'code': 200, 'message': 'snapshot', 'data': {'w': self.frame_width, 'h': self.frame_height, 'image': image_base64}}
    else:
      return {'code': 500, 'message': 'No frame received', 'data': {}}
