package main

import (
  "fmt"
  "time"
	"net/http"
  "github.com/gin-gonic/gin"
)


var gSessions_gin = make(map[string]*Session)

type SessionID_gin struct {
  Id string `json:"seesion_id"`
}

type Profile_gin struct {
  Name string `json:"profile"`
}

type Position_gin struct {
  Pan float64 `json:"Pan"`
  Tilt float64 `json:"Tilt"`
  Zoom float64 `json:"Zoom"`
  PanSpeed float64 `json:"PanSpeed"`
  TiltSpeed float64 `json:"TiltSpeed"`
  ZoomSpeed float64 `json:"ZoomSpeed"`
}


func checkGinSessionExpire() {
  for {
    for id, session:= range gSessions_gin {
      if session.session_end {
        session.image = nil
        session.lock = nil
        session = nil
        fmt.Println("Session expired: " + id)
        delete(gSessions_gin, id)
        break
      }
    }

    time.Sleep(1 * time.Second)
  }
}

func ApiHome(c *gin.Context) {
  c.JSON(http.StatusOK, gin.H{
    "message": "PTZ Server",
  })
}

func checkGinCookie(c *gin.Context) (string, error) {
  sid, err := c.Cookie("session_id")
  if err != nil {
    c.JSON(http.StatusUnauthorized, gin.H{
      "code": http.StatusUnauthorized,
      "message": "Unauthorized",
      "data": nil,
    })

    return "", err
  }

  return sid, nil
}

func Snapshot(c *gin.Context) {
  sid, err := checkGinCookie(c)

  if err != nil {
    return
  }

  json := gSessions_gin[sid].GetSnapshot()
  gSessions_gin[sid].ActivateSession()

  c.JSON(http.StatusOK, json)
}

func Connect(c *gin.Context) {
  info := PTZInfo{}

  if err := c.ShouldBindJSON(&info); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{
      "code": http.StatusBadRequest,
      "message": err.Error(),
      "data": nil,
    })
    return
  }

  found := false
  sid := ""
  for _, session:= range gSessions_gin {
    if session.ptz.info.Ip == info.Ip && session.ptz.info.Port == info.Port {
      found = true
      sid = session.id
      fmt.Println("Session exist: " + session.id)
      break
    }
  }

  if !found {
    session, err := NewSession(info.Ip, info.Port, info.Username, info.Password)

    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{
        "code": http.StatusInternalServerError,
        "message": err.Error(),
        "data": nil,
      })
      return
    } else {
      sid = session.id
      gSessions_gin[session.id] = session
    }
  }

  c.SetCookie("session_id", sid, 300, "/", "", false, true)

  c.JSON(http.StatusOK, gin.H{
    "code": http.StatusOK,
    "message": "Session started",
    "data": SessionID_gin{Id: sid},
  })
}

func GetConfigs(c *gin.Context) {
  sid, err := checkGinCookie(c)

  if err != nil {
    return
  }

  gSessions_gin[sid].ActivateSession()
  json, _ := gSessions_gin[sid].ptz.GetConfigs()

  c.JSON(http.StatusOK, json)
}

func GetPresets(c *gin.Context) {
  sid, err := checkGinCookie(c)

  if err != nil {
    return
  }

  gSessions_gin[sid].ActivateSession()
  json, _ := gSessions_gin[sid].ptz.GetPresets()

  c.JSON(http.StatusOK, json)
}

func GetPosition(c *gin.Context) {
  sid, err := checkGinCookie(c)

  if err != nil {
    return
  }

  gSessions_gin[sid].ActivateSession()
  json, _ := gSessions_gin[sid].ptz.GetPosition()

  c.JSON(http.StatusOK, json)
}

func IsMoving(c *gin.Context) {
  sid, err := checkGinCookie(c)

  if err != nil {
    return
  }

  gSessions_gin[sid].ActivateSession()
  json, _ := gSessions_gin[sid].ptz.IsMoving()

  c.JSON(http.StatusOK, json)
}

func ChangeProfile(c *gin.Context) {
  sid, err := checkGinCookie(c)

  if err != nil {
    return
  }

  profile := Profile_gin{}

  if err := c.ShouldBindJSON(&profile); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{
      "code": http.StatusBadRequest,
      "message": err.Error(),
      "data": nil,
    })
    return
  }

  gSessions_gin[sid].ActivateSession()
  if profile.Name != gSessions_gin[sid].ptz.profile_name {
    err = gSessions_gin[sid].ChangeProfile(profile.Name)

    if err != nil {
      c.JSON(http.StatusInternalServerError, gin.H{
        "code": http.StatusInternalServerError,
        "message": err.Error(),
        "data": nil,
      })
      return
    }
  }

  c.JSON(http.StatusInternalServerError, gin.H{
    "code": http.StatusOK,
    "message": "Change profile",
    "data": profile,
  })
}

func RelativeMove(c *gin.Context) {
  sid, err := checkGinCookie(c)

  if err != nil {
    return
  }

  pos := Position_gin{}

  if err := c.ShouldBindJSON(&pos); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{
      "code": http.StatusBadRequest,
      "message": err.Error(),
      "data": nil,
    })
    return
  }

  gSessions_gin[sid].ActivateSession()
  json, _ := gSessions_gin[sid].ptz.MoveRelativePosition(pos.Pan, pos.Tilt, pos.Zoom, pos.PanSpeed, pos.TiltSpeed, pos.ZoomSpeed)

  c.JSON(http.StatusOK, json)
}

func GotoPosition(c *gin.Context) {
  sid, err := checkGinCookie(c)

  if err != nil {
    return
  }

  pos := Position_gin{}

  if err := c.ShouldBindJSON(&pos); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{
      "code": http.StatusBadRequest,
      "message": err.Error(),
      "data": nil,
    })
    return
  }

  gSessions_gin[sid].ActivateSession()
  json, _ := gSessions_gin[sid].ptz.GotoPosition(pos.Pan, pos.Tilt, pos.Zoom, pos.PanSpeed, pos.TiltSpeed, pos.ZoomSpeed)

  c.JSON(http.StatusOK, json)
}

func GotoPreset(c *gin.Context) {
  sid, err := checkGinCookie(c)

  if err != nil {
    return
  }

  preset := PTZPresetID{}

  if err := c.ShouldBindJSON(&preset); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{
      "code": http.StatusBadRequest,
      "message": err.Error(),
      "data": nil,
    })
    return
  }

  gSessions_gin[sid].ActivateSession()
  json, _ := gSessions_gin[sid].ptz.GotoPreset(preset.Id)

  c.JSON(http.StatusOK, json)
}

func GotoHome(c *gin.Context) {
  sid, err := checkGinCookie(c)

  if err != nil {
    return
  }

  gSessions_gin[sid].ActivateSession()
  json, _ := gSessions_gin[sid].ptz.GotoHome()

  c.JSON(http.StatusOK, json)
}

func Stop(c *gin.Context) {
  sid, err := checkGinCookie(c)

  if err != nil {
    return
  }

  gSessions_gin[sid].ActivateSession()
  json, _ := gSessions_gin[sid].ptz.Stop()

  c.JSON(http.StatusOK, json)
}

func server_gin_main() {
  gin.SetMode(gin.ReleaseMode)

  router := gin.Default()
  
  router.StaticFile("/", "./static/index.html")
  router.StaticFile("/index.html", "./static/index.html")
  router.Static("/favicon.ico", "./static/favicon.ico")
  router.Static("/js", "./static/js")
  router.Static("/css", "./static/css")

  router.GET("/ptz", ApiHome)
  router.GET("/snapshot", Snapshot)
  router.GET("/ptz/config", GetConfigs)
  router.GET("/ptz/presets", GetPresets)
  router.GET("/ptz/position", GetPosition)
  router.GET("/ptz/moving", IsMoving)
  router.POST("/ptz/connect", Connect)
  router.POST("/ptz/profile", ChangeProfile)
  router.POST("/ptz/move/relative", RelativeMove)
  router.POST("/ptz/goto/position", GotoPosition)
  router.POST("/ptz/goto/preset", GotoPreset)
  router.POST("/ptz/goto/home", GotoHome)
  router.POST("/ptz/stop", Stop)

  fmt.Println("Starting Restful server on port 8000.")

  go checkGinSessionExpire()
  
  router.Run(":8000")
}


