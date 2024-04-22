package main

import (
	"fmt"
	"io"
	"errors"
	"net/http"
	"strconv"
	"context"
	goonvif "github.com/use-go/onvif"
	"github.com/use-go/onvif/media"
	"github.com/use-go/onvif/ptz"
	"github.com/use-go/onvif/xsd/onvif"
	"github.com/beevik/etree"
)

type PTZControl struct{
	ctx context.Context
	cam *goonvif.Device
	connected bool
	info PTZInfo
	configs PTZConfigs
	profiles map[string]string
	profile_name string
}

type PTZInfo struct {
	Ip string `json:"ip"`
	Port uint16 `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// type PTZResponse struct {
// 	code int
// 	message string
// 	data interface{}
// }

type Stream struct {
	Token string
	Name string
	Video struct {
		Encoding string
		Resolution struct {
			Width uint32
			Height uint32
		}
		Quality float64
		RateControl struct {
			FrameRateLimit uint32
			EncodingInterval uint32
			BitrateLimit uint32
		}
	}
	Audio struct {
		SampleRate uint32
		Bitrate	uint32
		Encoding string
	}
}

type PTZRange struct {
	Min float32
	Max float32
}

type PTZConfig struct {
	Pan PTZRange
	Tilt PTZRange
	Zoom PTZRange
	// Speed PTZRange
}

type PTZConfigs struct {
	Streams []Stream
	PTZ PTZConfig 
}

type PTZStatus struct {
	Pan float32
	Tilt float32
	Zoom float32
	Moving bool
	PTMoving string
	ZMoving string
}

type PTZPreset struct {
	Id string
	Name string
	PTZPosition struct {
		Pan float32
		Tilt float32
		Zoom float32
	}
}

type PTZUri struct {
	Uri string
}

type PTZPresetID struct {
	Id string `json:"preset"`
}

type Moving struct {
	Moving bool  `json:"Moving"`
	PanTilt string  `json:"PanTilt"`
	Zoom string  `json:"Zoom"`
}

// Internal

 func not_connected() map[string]interface{} {
	return map[string]interface{}{"code": 404, "message": "Cannot connect to IP Camera", "data": nil}
 }

func readResponse(resp *http.Response) string {
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func str2uint32(s string) uint32 {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return uint32(i)
}

func str2float32(s string) float32 {
	f, err := strconv.ParseFloat(s, 32)
	if err != nil {
		panic(err)
	}
	return float32(f)
}

func str2float64(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err)
	}
	return f
}

func emptyConfig() PTZConfigs {
	configs := PTZConfigs{Streams: nil, PTZ: emptyPTZConfig()}
	return configs
}
 
 func emptyPTZConfig() PTZConfig {
	empty := PTZRange{Min: 0, Max: 0}
	ptz := PTZConfig{Pan: empty, Tilt: empty, Zoom: empty/*, Speed: empty*/}
	return ptz
}

 func emptyStatus() PTZStatus {
	 status := PTZStatus{Moving: false, Pan: 0, Tilt: 0, Zoom: 0}
	 return status
}
 
func getMediaProfiles(dev *goonvif.Device) ([]Stream, map[string]string, string, error) {
	getProfiles := media.GetProfiles{}
	getProfilesResponseXML, err := dev.CallMethod(getProfiles)
	
	xml := readResponse(getProfilesResponseXML)
	// fmt.Println(xml)
	doc := etree.NewDocument()
	streams := make([]Stream, 0)
	profiles := make(map[string]string, 0)
	name := ""
	if err := doc.ReadFromString(xml); err == nil {
		nodes := doc.Root().FindElements("/Envelope/Body/GetProfilesResponse/Profiles")

		for _, node := range nodes {
			var stream Stream
			stream.Token = node.SelectAttr("token").Value
			stream.Name = node.SelectElement("Name").Text()
			name = stream.Name
			profiles[stream.Name] = stream.Token
			stream.Video.Encoding = node.FindElement("VideoEncoderConfiguration/Encoding").Text()
			stream.Video.Quality = str2float64(node.FindElement("VideoEncoderConfiguration/Quality").Text())
			stream.Video.Resolution.Width = str2uint32(node.FindElement("VideoEncoderConfiguration/Resolution/Width").Text())
			stream.Video.Resolution.Height = str2uint32(node.FindElement("VideoEncoderConfiguration/Resolution/Height").Text())
			stream.Video.RateControl.FrameRateLimit = str2uint32(node.FindElement("VideoEncoderConfiguration/RateControl/FrameRateLimit").Text())
			stream.Video.RateControl.EncodingInterval = str2uint32(node.FindElement("VideoEncoderConfiguration/RateControl/EncodingInterval").Text())
			stream.Video.RateControl.BitrateLimit = str2uint32(node.FindElement("VideoEncoderConfiguration/RateControl/BitrateLimit").Text())
			stream.Audio.Bitrate = str2uint32(node.FindElement("AudioEncoderConfiguration/Bitrate").Text())
			stream.Audio.Encoding = node.FindElement("AudioEncoderConfiguration/Encoding").Text()
			stream.Audio.SampleRate = str2uint32(node.FindElement("AudioEncoderConfiguration/SampleRate").Text())

			streams = append(streams, stream)
		}

		return streams, profiles, name, nil
	}

	return streams, profiles, name, err
}

func getPTZProfile(dev *goonvif.Device) (PTZConfig, error) {
	getConfigurations := ptz.GetConfigurations{}
	getConfigurationsResponseXML, err := dev.CallMethod(getConfigurations)
	
	xml := readResponse(getConfigurationsResponseXML)
	// fmt.Println(xml)
	doc := etree.NewDocument()

	if err := doc.ReadFromString(xml); err == nil {
		config := doc.Root().FindElement("/Envelope/Body/GetConfigurationsResponse/PTZConfiguration")
		// token := config.SelectAttr("token").Value
		var pan PTZRange
		var tilt PTZRange
		var zoom PTZRange
		pan.Min = str2float32(config.FindElement("PanTiltLimits/Range/XRange/Min").Text())
		pan.Max = str2float32(config.FindElement("PanTiltLimits/Range/XRange/Max").Text())
		tilt.Min = str2float32(config.FindElement("PanTiltLimits/Range/YRange/Min").Text())
		tilt.Max = str2float32(config.FindElement("PanTiltLimits/Range/YRange/Max").Text())
		zoom.Min = str2float32(config.FindElement("ZoomLimits/Range/XRange/Min").Text())
		zoom.Max = str2float32(config.FindElement("ZoomLimits/Range/XRange/Max").Text())

		// speed, err := getPTZSpeed(dev, token)

		// if err == nil {
			ptz := PTZConfig {
				Pan: pan,
				Tilt: tilt,
				Zoom: zoom,
				// Speed: speed,
			}
			return ptz, nil
		// }
	}

	return emptyPTZConfig(), err
}

// func getPTZSpeed(dev *goonvif.Device, token string) (PTZRange, error) {
// 	getConfigurationOptions := ptz.GetConfigurationOptions{ProfileToken: onvif.ReferenceToken(token)}
// 	getConfigurationOptionsResponseXML, err := dev.CallMethod(getConfigurationOptions)
	
// 	xml := readResponse(getConfigurationOptionsResponseXML)

// 	fmt.Println(xml)
// 	doc := etree.NewDocument()

// 	if err := doc.ReadFromString(xml); err == nil {
// 		space := doc.Root().FindElement("/Envelope/Body/GetConfigurationOptionsResponse/PTZConfigurationOptions/Spaces/ZoomSpeedSpace")

// 		var speed PTZRange

// 		speed.Min = str2float32(space.FindElement("Range/XRange/Min").Text())
// 		speed.Max = str2float32(space.FindElement("Range/XRange/Max").Text())

// 		return speed, nil
// 	}

// 	empty := PTZRange{Min: 0, Max: 0}
// 	return empty, err
// }

// func getProfilesSDK(ctx *context.Context, dev *goonvif.Device) ([]Stream, error) {
// 	getProfiles := media.GetProfiles{}
// 	getProfilesResponse, err := sdkmedia.Call_GetProfiles(*ctx, dev, getProfiles)

// 	streams := make([]Stream, 0)
// 	if err == nil {
// 		profiles := getProfilesResponse.Profiles
// 		for profile := range profiles {
// 			var stream Stream
// 			stream.Token = string(profiles[profile].Token)
// 			stream.Name = string(profiles[profile].Name)
// 			stream.Video.Encoding = string(profiles[profile].VideoEncoderConfiguration.Encoding)
// 			stream.Video.Quality = profiles[profile].VideoEncoderConfiguration.Quality
// 			stream.Video.Resolution.Width = uint32(profiles[profile].VideoEncoderConfiguration.Resolution.Width)
// 			stream.Video.Resolution.Height = uint32(profiles[profile].VideoEncoderConfiguration.Resolution.Height)
// 			stream.Video.RateControl.FrameRateLimit = uint32(profiles[profile].VideoEncoderConfiguration.RateControl.FrameRateLimit)
// 			stream.Video.RateControl.EncodingInterval = uint32(profiles[profile].VideoEncoderConfiguration.RateControl.EncodingInterval)
// 			stream.Video.RateControl.BitrateLimit = uint32(profiles[profile].VideoEncoderConfiguration.RateControl.BitrateLimit)
// 			stream.Audio.Bitrate = uint32(profiles[profile].AudioEncoderConfiguration.Bitrate)
// 			stream.Audio.Encoding = string(profiles[profile].AudioEncoderConfiguration.Encoding)
// 			stream.Audio.SampleRate = uint32(profiles[profile].AudioEncoderConfiguration.SampleRate)

// 			streams = append(streams, stream)
// 		}

// 		return streams, nil
// 	}

// 	return streams, err
// }

func getStatus(dev *goonvif.Device) (PTZStatus, error) {
	getStatus := ptz.GetStatus{}
	getStatusResponseXML, err := dev.CallMethod(getStatus)
	
	xml := readResponse(getStatusResponseXML)
	// fmt.Println(xml)
	doc := etree.NewDocument()

	if err := doc.ReadFromString(xml); err == nil {
		var status PTZStatus
		node := doc.Root().FindElement("/Envelope/Body/GetStatusResponse/PTZStatus")
		pt := node.FindElement("Position/PanTilt")

		status.Pan = str2float32(pt.SelectAttr("x").Value)
		status.Tilt = str2float32(pt.SelectAttr("y").Value)

		z := node.FindElement("Position/Zoom")
		status.Zoom = str2float32(z.SelectAttr("x").Value)

		status.PTMoving = node.FindElement("MoveStatus/PanTilt").Text()
		status.ZMoving = node.FindElement("MoveStatus/Zoom").Text()

		status.Moving = (status.PTMoving == "moving" || status.ZMoving == "moving")

		return status, nil
	}

	return emptyStatus(), err
}

func getPresets(dev *goonvif.Device, token string) ([]PTZPreset, error) {
	getPresets := ptz.GetPresets{ProfileToken: onvif.ReferenceToken(token)}
	getPresetsResponseXML, err := dev.CallMethod(getPresets)
	
	xml := readResponse(getPresetsResponseXML)
	// fmt.Println(xml)
	doc := etree.NewDocument()
	presets := make([]PTZPreset, 0)

	if err := doc.ReadFromString(xml); err == nil {
		var preset PTZPreset
		nodes := doc.Root().FindElements("/Envelope/Body/GetPresetsResponse/Preset")
		
		for _, node := range nodes {
			preset.Id = node.SelectAttr("token").Value
			preset.Name = node.SelectElement("Name").Text()
			preset.PTZPosition.Pan = str2float32(node.FindElement("PTZPosition/PanTilt").SelectAttr("x").Value)
			preset.PTZPosition.Tilt = str2float32(node.FindElement("PTZPosition/PanTilt").SelectAttr("y").Value)
			preset.PTZPosition.Zoom = str2float32(node.FindElement("PTZPosition/Zoom").SelectAttr("x").Value)

			presets = append(presets, preset)
		}
		return presets, nil
	}

	return nil, err
}

func getStreamUri(dev *goonvif.Device, token string) (string, error) {
	transport := onvif.Transport{Protocol: onvif.TransportProtocol("RTSP"), Tunnel: nil}
	setup := onvif.StreamSetup{Stream: onvif.StreamType("RTP-Unicast"), Transport: transport}
	getStreamUri := media.GetStreamUri{ProfileToken: onvif.ReferenceToken(token), StreamSetup: setup}
	getStreamUriResponseXML, err := dev.CallMethod(getStreamUri)
	
	xml := readResponse(getStreamUriResponseXML)
	// fmt.Println(xml)
	doc := etree.NewDocument()

	if err := doc.ReadFromString(xml); err == nil {
		uri := doc.Root().FindElement("/Envelope/Body/GetStreamUriResponse/MediaUri/Uri").Text()

		return uri, nil
	}

	return "", err
}

func stop(dev *goonvif.Device, token string) (error) {
	stop := ptz.Stop{ProfileToken: onvif.ReferenceToken(token), PanTilt: true, Zoom: true}
	stopResponseXML, err := dev.CallMethod(stop)
	
	xml := readResponse(stopResponseXML)
	// fmt.Println(xml)
	doc := etree.NewDocument()

	if err := doc.ReadFromString(xml); err == nil {
		res := doc.Root().FindElement("/Envelope/Body/StopResponse")

		if res != nil {
			return nil
		}

		return errors.New("stopResponse not found")
	}

	return err
}

func gotoPreset(dev *goonvif.Device, token string, id string) (string, error) {
	gotoPreset := ptz.GotoPreset{ProfileToken: onvif.ReferenceToken(token), PresetToken: onvif.ReferenceToken(id)}
	gotoPresetResponseXML, err := dev.CallMethod(gotoPreset)
	
	xml := readResponse(gotoPresetResponseXML)
	// fmt.Println(xml)
	doc := etree.NewDocument()

	if err := doc.ReadFromString(xml); err == nil {
		res := doc.Root().FindElement("/Envelope/Body/GotoPresetResponse")

		if res != nil {
			return id, nil
		}

		return id, errors.New("gotoPresetResponse not found")
	}

	return id, err
}

func gotoPosition(dev *goonvif.Device, token string, p float64, t float64, z float64, ps float64, ts float64, zs float64) (error) {
	position := onvif.PTZVector{PanTilt: onvif.Vector2D{X: p, Y: t}, Zoom: onvif.Vector1D{X: z}}
	speed := onvif.PTZSpeed{PanTilt: onvif.Vector2D{X: ps, Y: ts}, Zoom: onvif.Vector1D{X: zs}}
	gotoPosition := ptz.AbsoluteMove{ProfileToken: onvif.ReferenceToken(token), Position: position, Speed: speed}
	gotoPositionResponseXML, err := dev.CallMethod(gotoPosition)
	
	xml := readResponse(gotoPositionResponseXML)
	// fmt.Println(xml)
	doc := etree.NewDocument()

	if err := doc.ReadFromString(xml); err == nil {
		res := doc.Root().FindElement("/Envelope/Body/AbsoluteMoveResponse")

		if res != nil {
			return nil
		}

		return errors.New("absoluteMoveResponse not found")
	}

	return err
}


func moveRelativePosition(dev *goonvif.Device, token string, p float64, t float64, z float64, ps float64, ts float64, zs float64) (error) {
	position := onvif.PTZVector{PanTilt: onvif.Vector2D{X: p, Y: t}, Zoom: onvif.Vector1D{X: z}}
	speed := onvif.PTZSpeed{PanTilt: onvif.Vector2D{X: ps, Y: ts}, Zoom: onvif.Vector1D{X: zs}}
	gotoPosition := ptz.RelativeMove{ProfileToken: onvif.ReferenceToken(token), Translation: position, Speed: speed}
	gotoPositionResponseXML, err := dev.CallMethod(gotoPosition)
	
	xml := readResponse(gotoPositionResponseXML)
	// fmt.Println(xml)
	doc := etree.NewDocument()

	if err := doc.ReadFromString(xml); err == nil {
		res := doc.Root().FindElement("/Envelope/Body/RelativeMoveResponse")

		if res != nil {
			return nil
		}

		return errors.New("relativeResponse not found")
	}

	return err
}

// Interface for Outside

func NewPTZControl(ip string, port uint16, username string, password string) (*PTZControl, error) {
	ctx := context.Background()
	ptzInfo := PTZInfo{Ip: ip, Port: port, Username: username, Password: password}
	dev, err := goonvif.NewDevice(goonvif.DeviceParams{Xaddr: fmt.Sprintf("%s:%d", ip, port), Username: username, Password: password})
	if err == nil {
		streams, profiles, name, err := getMediaProfiles(dev)
		// profiles, err := getProfilesSDK(&ctx, dev)

		if err == nil {
			ptz, err := getPTZProfile(dev)

			if err == nil {
				configs := PTZConfigs{
					Streams: streams,
					PTZ: ptz,
				}

				return &PTZControl{
					ctx: ctx,
					cam: dev,
					connected: true,
					info: ptzInfo,
					configs: configs,
					profiles: profiles,
					profile_name: name,
				}, nil
			}
		}
	}

	return &PTZControl{
		ctx: ctx,
		cam: nil,
		connected: false,
		info: ptzInfo,
		configs: emptyConfig(),
		profiles: nil,
		profile_name: "",
	}, err

}

func (ptz *PTZControl) SetProfile(name string) {
	_, ok := ptz.profiles[name]
	if ok {
		ptz.profile_name = name
	}
}

func (ptz *PTZControl) GetConfigs() (map[string]interface{}, error) {
	if !ptz.connected {
		return not_connected(), errors.New("not connected")
	}

	return map[string]interface{}{"code": 200, "message": "PTZ config", "data": ptz.configs}, nil
}

func (ptz *PTZControl) IsMoving() (map[string]interface{}, error) {
	if !ptz.connected {
		return not_connected(), errors.New("not connected")
	}

	status, err := getStatus(ptz.cam)

	if err != nil {
		return map[string]interface{}{"code": 404, "message": "Cannot get PTZ status", "data": nil}, err
	}
	
	data := Moving{Moving:status.Moving, PanTilt: status.PTMoving, Zoom: status.ZMoving}

	return map[string]interface{}{"code": 200, "message": "PTZ moving status", "data": data}, err
}

func (ptz *PTZControl) GetPosition() (map[string]interface{}, error) {
	if !ptz.connected {
		return not_connected(), errors.New("not connected")
	}

	status, err := getStatus(ptz.cam)

	if err != nil {
		return map[string]interface{}{"code": 404, "message": "Cannot get PTZ status", "data": nil}, err
	}
	
	return map[string]interface{}{"code": 200, "message": "PTZ position", "data": status}, nil
}

func (ptz *PTZControl) GetPresets() (map[string]interface{}, error) {
	if !ptz.connected {
		return not_connected(), errors.New("not connected")
	}

	presets, err := getPresets(ptz.cam, ptz.profiles[ptz.profile_name])

	data := map[string]interface{}{"Presets": presets}

	if err != nil {
		return map[string]interface{}{"code": 404, "message": "Cannot get PTZ presets", "data": data}, err
	}
	
	return map[string]interface{}{"code": 200, "message": "PTZ presets", "data": data}, nil
}

func (ptz *PTZControl) GetStreamUri() (map[string]interface{}, error) {
	if !ptz.connected {
		return not_connected(), errors.New("not connected")
	}

	uri, err := getStreamUri(ptz.cam, ptz.profiles[ptz.profile_name])

	if err != nil {
		return map[string]interface{}{"code": 404, "message": "Cannot get Stream URI", "data": nil}, err
	}
	
	return map[string]interface{}{"code": 200, "message": "Stream URI", "data": PTZUri{Uri: uri}}, nil
}

func (ptz *PTZControl) GotoPreset(id string) (map[string]interface{}, error) {
	if !ptz.connected {
		return not_connected(), errors.New("not connected")
	}

	token, err := gotoPreset(ptz.cam, ptz.profiles[ptz.profile_name], id)

	if err != nil {
		return map[string]interface{}{"code": 404, "message": "Cannot set the camera to preset position", "data": nil}, err
	}
	
	return map[string]interface{}{"code": 200, "message": "Set the camera to preset position", "data": PTZPresetID{Id: token}}, nil
}

func (ptz *PTZControl) Stop() (map[string]interface{}, error) {
	if !ptz.connected {
		return not_connected(), errors.New("not connected")
	}

	err := stop(ptz.cam, ptz.profiles[ptz.profile_name])

	if err != nil {
		return map[string]interface{}{"code": 404, "message": "Cannot stop PTZ movement", "data": nil}, err
	}
	
	return map[string]interface{}{"code": 200, "message": "Stop PTZ", "data": nil}, nil
}

func (ptz *PTZControl) GotoPosition(p float64, t float64, z float64, ps float64, ts float64, zs float64) (map[string]interface{}, error) {
	if !ptz.connected {
		return not_connected(), errors.New("not connected")
	}

	err := gotoPosition(ptz.cam, ptz.profiles[ptz.profile_name], p, t, z, ps, ts, zs)

	if err != nil {
		return map[string]interface{}{"code": 404, "message": "Cannot set PTZ to position", "data": nil}, err
	}
	
	return map[string]interface{}{"code": 200, "message": "Set PTZ to position", "data": nil}, nil
}

func (ptz *PTZControl) GotoHome() (map[string]interface{}, error) {
	if !ptz.connected {
		return not_connected(), errors.New("not connected")
	}

	err := gotoPosition(ptz.cam, ptz.profiles[ptz.profile_name], 0, 0, 0, 1, 1, 1)

	if err != nil {
		return map[string]interface{}{"code": 404, "message": "Cannot set PTZ to Home position", "data": nil}, err
	}
	
	return map[string]interface{}{"code": 200, "message": "Set PTZ to Home position", "data": nil}, nil
}

func (ptz *PTZControl) MoveRelativePosition(p float64, t float64, z float64, ps float64, ts float64, zs float64) (map[string]interface{}, error) {
	if !ptz.connected {
		return not_connected(), errors.New("not connected")
	}

	err := moveRelativePosition(ptz.cam, ptz.profiles[ptz.profile_name], p, t, z, ps, ts, zs)

	if err != nil {
		return map[string]interface{}{"code": 404, "message": "Cannot set PTZ to relative position", "data": nil}, err
	}
	
	return map[string]interface{}{"code": 200, "message": "Set PTZ to relative position", "data": nil}, nil
}
