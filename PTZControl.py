from onvif import ONVIFCamera

class PTZControl:
  def __init__(self, ip, port, username, password):
    self.ip = ip
    self.port = port
    self.username = username
    self.password = password
    self.cam = None
    self.ptz = None
    self.media = None
    self.media_profiles = None
    self.profile_token_dict = dict()
    self.ptz_configs = None
    self.ptz_option = None
    self.ptz_presets = None
    self.profile = ''

    self.connected = self._connect()

    self._get_ptz_option()

  def _connect(self):
    try:
      self.cam = ONVIFCamera(self.ip, self.port, self.username, self.password)
      self.media = self.cam.create_media_service()
      self.ptz = self.cam.create_ptz_service()

      self.media_profiles = self.media.GetProfiles()  # 获取配置信息

      for profile in self.media_profiles:
        self.profile_token_dict[profile.Name] = profile.token
        self.profile = profile.Name

      return True
    except:
      return False
  
  def is_connected(self):
    return self.connected
  
  def _not_connected(self):
    return {'code': 404, 'message': 'Cannot connect to IP Camera', 'data': {}}

  def _get_ptz_option(self):
    if not self.connected:
      return self._not_connected()


    # for config in self.ptz_configs:
    if len(self.media_profiles) > 0:
      config_token = self.media_profiles[0].PTZConfiguration["token"]
      params = self.ptz.create_type('GetConfigurationOptions')
      params.ConfigurationToken = config_token
      self.ptz_option = self.ptz.GetConfigurationOptions(params)

  def _get_status(self, profile_id=0):
    if not self.connected:
      return None
    
    if profile_id >= len(self.media_profiles):
      return None
    
    params = self.ptz.create_type('GetStatus')
    params.ProfileToken = self.media_profiles[profile_id].token
    status = self.ptz.GetStatus(params)

    return status

  def get_configs(self):
    if not self.connected:
      return self._not_connected()
    
    configs = []

    for profile in self.media_profiles:
      config = {
        'Name': profile.Name,
        'Video': {
          'Encoding': profile.VideoEncoderConfiguration.Encoding,
          'Resolution': {
            'Width': profile.VideoEncoderConfiguration.Resolution.Width,
            'Height': profile.VideoEncoderConfiguration.Resolution.Height
          },
          'Quality': profile.VideoEncoderConfiguration.Quality,
          'RateControl': {
            'FrameRateLimit': profile.VideoEncoderConfiguration.RateControl.FrameRateLimit,
            'EncodingInterval': profile.VideoEncoderConfiguration.RateControl.EncodingInterval,
            'BitrateLimit': profile.VideoEncoderConfiguration.RateControl.BitrateLimit
          }
        },
        'Audio': {
          'SampleRate': profile.AudioEncoderConfiguration.SampleRate,
          'Bitrate': profile.AudioEncoderConfiguration.Bitrate,
          'Encoding': profile.AudioEncoderConfiguration.Encoding
        }
      }

      configs.append(config)
          
    ptz = {}

    if self.ptz_option is not None:
      ptz = {
        'Pan': {
          'Min': self.ptz_option.Spaces.AbsolutePanTiltPositionSpace[0].XRange.Min,
          'Max': self.ptz_option.Spaces.AbsolutePanTiltPositionSpace[0].XRange.Max,
        },
        'Tilt': {
          'Min': self.ptz_option.Spaces.AbsolutePanTiltPositionSpace[0].YRange.Min,
          'Max': self.ptz_option.Spaces.AbsolutePanTiltPositionSpace[0].YRange.Max,
        },
        'Speed': {
          'Min': self.ptz_option.Spaces.PanTiltSpeedSpace[0].XRange.Min,
          'Max': self.ptz_option.Spaces.PanTiltSpeedSpace[0].XRange.Max,
        }
      }

    return {'code': 200, 'message': 'PTZ config', 'data': {'Streams': configs, 'PTZ': ptz}}
    
  def get_position(self):
    if not self.connected:
      return self._not_connected()
    
    status = self._get_status()

    if status == None:
      return {'code': 404, 'message': 'Cannot get PTZ status', 'data': {}}

    moving = True if status.Position.PanTilt.x == 'moving' or status.MoveStatus.Zoom == 'moving' else False
    
    return {'code': 200, 'message': 'PTZ position', 'data': {'Moving': moving, 'Pan': status.Position.PanTilt.x, 'Tilt': status.Position.PanTilt.y, 'Zoom': status.Position.Zoom.x}}
 
  def goto_home(self):
    if not self.connected:
      return self.not_connected()
    
    self.goto_position(0, 0, 0)

    return {'code': 200, 'message': 'Set PTZ to home position', 'data': {}}
  
  def goto_position(self, pan=0, tilt=0, zoom=0, pan_speed=1, tilt_speed=1, zoom_speed=1):
    if not self.connected:
      return self._not_connected()
    
    params = {'ProfileToken': self.profile_token_dict[self.profile], 'Position': {'PanTilt': {'x': pan, 'y': tilt}, 'Zoom': zoom}, 'Speed': {'PanTilt': {'x': pan_speed, 'y': tilt_speed}, 'Zoom': zoom_speed}}
    self.ptz.AbsoluteMove(params)
    return {'code': 200, 'message': 'Set PTZ to position', 'data': {}}

  def move_relative_position(self, pan=0.1, tilt=0.1, zoom=0, pan_speed=1, tilt_speed=1, zoom_speed=1):
    if not self.connected:
      return self._not_connected()

    params = {'ProfileToken': self.profile_token_dict[self.profile], 'Translation': {'PanTilt': {'x': pan, 'y': tilt}, 'Zoom': zoom}, 'Speed': {'PanTilt': {'x': pan_speed, 'y': tilt_speed}, 'Zoom': zoom_speed}}
    self.ptz.RelativeMove(params)
    return {'code': 200, 'message': 'Set PTZ to relative position', 'data': {}}

  def get_presets(self):
    if not self.connected:
      return self._not_connected()
    
    params = self.ptz.create_type('GetPresets')
    params.ProfileToken = self.profile_token_dict[self.profile]
    # params.ProfileToken = 'profile_1'
    self.ptz_presets = self.ptz.GetPresets(params)

    presets = []
    for preset in self.ptz_presets:
      position = {
        'Id': preset.token,
        'Name': preset.Name,
        'PTZPosition': {
          'Pan': preset.PTZPosition.PanTilt.x,
          'Tilt': preset.PTZPosition.PanTilt.y,
          'Zoom': preset.PTZPosition.Zoom.x
        }
      }

      presets.append(position)

    return {'code': 200, 'message': 'PTZ presets', 'data': { 'Presets': presets }}
    
  def goto_preset(self, preset_id):
    if not self.connected:
      return self._not_connected()
    
    params = self.ptz.create_type('GotoPreset')
    params.ProfileToken = self.profile_token_dict[self.profile]
    params.PresetToken = preset_id
    self.ptz.GotoPreset(params)

    return {'code': 200, 'message': 'Set the camera to preset position', 'data': {'Id': preset_id}}
  
  def ptz_stop(self):
    if not self.connected:
      return self._not_connected()
    
    params = self.ptz.create_type('Stop')
    params.ProfileToken = self.profile_token_dict[self.profile]
    params.PanTilt = True
    params.Zoom = True
    self.ptz.Stop(params)
    return {'code': 200, 'message': 'Stop PTZ', 'data': {}}

  def get_stream_uri(self):
    if not self.connected:
      return self._not_connected()
    
    params = self.media.create_type('GetStreamUri')
    params.ProfileToken = self.profile_token_dict[self.profile]
    params.StreamSetup = {'Stream': 'RTP-Unicast', 'Transport': 'RTSP'}
    uri = self.media.GetStreamUri(params)

    return {'code': 200, 'message': 'Stream URI', 'data': {'Uri': uri.Uri}}
  
  def set_profile(self, profile):
    self.profile = profile
    return {'code': 200, 'message': 'Change profile', 'data': {'profile': profile}}
    
  def is_moving(self):
    if not self.connected:
      return self._not_connected()
    
    status = self._get_status()

    if status == None:
      return {'code': 404, 'message': 'Cannot get PTZ status', 'data': {}}

    moving = True if status.MoveStatus.PanTilt == 'moving' or status.MoveStatus.Zoom == 'moving' else False
    
    # posible value for PanTilt and Zoom: idle, moving
    return {'code': 200, 'message': 'PTZ moving status', 'data': {'Moving': moving, 'PanTilt': status.MoveStatus.PanTilt, 'Zoom': status.MoveStatus.Zoom}}
    
      
