package main

import (
	"fmt"
	"time"
	"bytes"
	"strings"
	"strconv"
	"log"
	"image"
	"image/jpeg"
	// "errors"
	"sync"
  "encoding/base64"
	"github.com/google/uuid"
	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/base"
	"github.com/bluenviron/gortsplib/v4/pkg/format"
	"github.com/bluenviron/gortsplib/v4/pkg/format/rtph264"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"github.com/bluenviron/gortsplib/v4/pkg/format/rtph265"
	"github.com/bluenviron/mediacommon/pkg/codecs/h265"
	"github.com/pion/rtp"
)

type Session struct {
	id string
	ptz *PTZControl
  last_time time.Time
	session_end bool
	stop_video bool
	video_stopped bool
	image image.Image
	lock *sync.RWMutex
}

const gTimeout = 30

// Internal threads

// Check session timeout
func checkSession(session *Session) {
	for {
		now := time.Now()

		if session.last_time.Add(time.Second * gTimeout).Before(now) {
			// fmt.Println("timeout")
			session.stop_video = true
			break
		}
		time.Sleep(1 * time.Second)
	}

	fmt.Println("Session timeout: " + session.id)
	session.session_end = true
}

func processStream(uri string, session *Session) {
	// test loop
	// for {
	// 	if session.stop_video {
	// 		break
	// 	}

	// 	time.Sleep(1 * time.Second)
	// }

	// fmt.Println("Streaming End: " + session.ptz.info.Ip + ":" + strconv.Itoa(int(session.ptz.info.Port)))

	is_h264 := true
	c := gortsplib.Client{}

	fmt.Println(uri)
	
	// parse URL
	u, err := base.ParseURL(uri)
	if err != nil {
		panic(err)
	}

	// connect to the server
	err = c.Start(u.Scheme, u.Host)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	// find available medias
	desc, _, err := c.Describe(u)
	if err != nil {
		panic(err)
	}

	var forma264 *format.H264
	var forma265 *format.H265

	// find the H264 media and formatv
	medi := desc.FindFormat(&forma264)
	if medi == nil {
		// find the H265 media and format
		medi = desc.FindFormat(&forma265)
		if medi == nil {
			panic("media not found")
		}
		is_h264 = false
	}

	// setup RTP/H264 -> H264 decoder
	if is_h264 {
		rtpDec, err := forma264.CreateDecoder()
		if err != nil {
			panic(err)
		}

		// setup H264 -> raw frames decoder
		frameDec := &h264Decoder{}
		err = frameDec.initialize()
		if err != nil {
			panic(err)
		}
		defer frameDec.close()

		// if SPS and PPS are present into the SDP, send them to the decoder
		if forma264.SPS != nil {
			frameDec.decode(forma264.SPS)
		}
		if forma264.PPS != nil {
			frameDec.decode(forma264.PPS)
		}

		// setup a single media
		_, err = c.Setup(desc.BaseURL, medi, 0, 0)
		if err != nil {
			panic(err)
		}

		iframeReceived := false

		// called when a RTP packet arrives
		c.OnPacketRTP(medi, forma264, func(pkt *rtp.Packet) {
			// extract access units from RTP packets
			au, err := rtpDec.Decode(pkt)
			if err != nil {
				if err != rtph264.ErrNonStartingPacketAndNoPrevious && err != rtph264.ErrMorePacketsNeeded {
					log.Printf("ERR: %v", err)
				}
				return
			}

			// wait for an I-frame
			if !iframeReceived {
				if !h264.IDRPresent(au) {
					log.Printf("waiting for an I-frame")
					return
				}
				iframeReceived = true
			}

			for _, nalu := range au {
				// convert NALUs into RGBA frames
				img, _ := frameDec.decode(nalu)
				// if err != nil {
				// 	panic(err)
				// }

				// wait for a frame
				if img == nil {
					continue
				}

				session.image = img
			}
		})
	} else {
		// setup RTP/H265 -> H265 decoder
		rtpDec, err := forma265.CreateDecoder()
		if err != nil {
			panic(err)
		}

		// setup H265 -> raw frames decoder
		frameDec := &h265Decoder{}
		err = frameDec.initialize()
		if err != nil {
			panic(err)
		}
		defer frameDec.close()

		// if VPS, SPS and PPS are present into the SDP, send them to the decoder
		if forma265.VPS != nil {
			frameDec.decode(forma265.VPS)
		}
		if forma265.SPS != nil {
			frameDec.decode(forma265.SPS)
		}
		if forma265.PPS != nil {
			frameDec.decode(forma265.PPS)
		}

		// setup a single media
		_, err = c.Setup(desc.BaseURL, medi, 0, 0)
		if err != nil {
			panic(err)
		}

		iframeReceived := false

		// called when a RTP packet arrives
		c.OnPacketRTP(medi, forma265, func(pkt *rtp.Packet) {
			// extract access units from RTP packets
			au, err := rtpDec.Decode(pkt)
			if err != nil {
				if err != rtph265.ErrNonStartingPacketAndNoPrevious && err != rtph265.ErrMorePacketsNeeded {
					log.Printf("ERR: %v", err)
				}
				return
			}

			// wait for an I-frame
			if !iframeReceived {
				if !h265.IsRandomAccess(au) {
					log.Printf("waiting for an I-frame")
					return
				}
				iframeReceived = true
			}

			for _, nalu := range au {
				// convert NALUs into RGBA frames
				img, err := frameDec.decode(nalu)
				if err != nil {
					panic(err)
				}

				// wait for a frame
				if img == nil {
					continue
				}

				session.lock.Lock()
				session.image = img
				session.lock.Unlock()
			}
		})
	}

	// start playing
	_, err = c.Play(nil)
	if err != nil {
		panic(err)
	}

	for {
		if session.stop_video {
			fmt.Println("Streaming End: " + session.ptz.info.Ip + ":" + strconv.Itoa(int(session.ptz.info.Port)))
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	session.video_stopped = true

	// wait until a fatal error
	// panic(c.Wait())
}

// get video stream and decode to Image
// func getVideo(uri string, session *Session) {

// 	defer func() {
// 		if err:=recover(); err!=nil{
// 			fmt.Println(err)
// 			session.stop_video = false
// 		}
// 	}()

// 	processStream(uri, session)

// 	session.session_end = true
// }

func NewSession(ip string, port uint16, username string, password string) (*Session, error) {
	ptz, err := NewPTZControl(ip, port, username, password)
	if err != nil {
		fmt.Println("init session error:", err)
		return &Session{}, err
	}

	res, _ := ptz.GetStreamUri()

	rtsp_uri := res["data"].(PTZUri).Uri

	rtsp_uri = strings.Replace(rtsp_uri, "rtsp://", "rtsp://"+ username + ":"+ password + "@", 1)

	uuid := uuid.New()

	session := Session{
		id: uuid.String(),
		ptz: ptz,
		last_time: time.Now(),
		session_end: false,
		stop_video: false,
		video_stopped: false,
		image: nil,
		lock: new(sync.RWMutex),
	}

	// Start video streaming thread
	go processStream(rtsp_uri, &session)

	// Start session timeout checking thread
	go checkSession(&session)

	fmt.Println("Session start: " + session.id + " - " + session.ptz.info.Ip + ":" + strconv.Itoa(int(session.ptz.info.Port)))

	return &session, nil
}

// Interface

func (session *Session) ActivateSession() {
	session.last_time = time.Now()
}

func (session *Session) ChangeProfile(profile string) error {
	// fmt.Println("Change profile: " + profile)
	session.ptz.SetProfile(profile)

	//Stop video stream
	session.stop_video = true

	res, err := session.ptz.GetStreamUri()

	if err != nil {
		return err
	}

	// Wait for video streaming stop
	for {
		if session.video_stopped {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	rtsp_uri := res["data"].(PTZUri).Uri

	rtsp_uri = strings.Replace(rtsp_uri, "rtsp://", "rtsp://"+ session.ptz.info.Username + ":"+ session.ptz.info.Password + "@", 1)

	session.stop_video = false
	session.video_stopped = false

	// Start video streaming thread
	go processStream(rtsp_uri, session)

	return nil
}

func (session *Session) GetSnapshot() map[string]interface{} {
	// fmt.Println("Snapshot")
	session.lock.RLock()

	if session.image == nil {
		session.lock.RUnlock()
		return map[string]interface{}{"code": 500, "message": "No frame received", "data": nil}
	}

	var buf bytes.Buffer
	size := session.image.Bounds().Size()

	err := jpeg.Encode(&buf, session.image, &jpeg.Options{
		Quality: 80,
	})

	session.lock.RUnlock()

	if err != nil {
		return map[string]interface{}{"code": 500, "message": "No frame received", "data": nil}
	}

	// fmt.Println(size)

	image_base64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	
	return map[string]interface{}{"code": 200, "message": "Snapshot", "data": map[string]interface{}{"w": size.X, "h": size.Y, "image": "data:image/jpeg;base64," + image_base64}}
}