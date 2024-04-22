package main

import (
  "os"
  "fmt"
  "mime"
  "time"
  "path/filepath"
	"net/http"
  "encoding/json"
)


var gSessions = make(map[string]*Session)

type SessionID struct {
  Id string `json:"seesion_id"`
}

type Profile struct {
  Name string `json:"profile"`
}

type Position struct {
  Pan float64 `json:"Pan"`
  Tilt float64 `json:"Tilt"`
  Zoom float64 `json:"Zoom"`
  PanSpeed float64 `json:"PanSpeed"`
  TiltSpeed float64 `json:"TiltSpeed"`
  ZoomSpeed float64 `json:"ZoomSpeed"`
}


func handleApiHome(w http.ResponseWriter, r *http.Request) {
  if r.Method != "GET" {
    w.WriteHeader(http.StatusNotFound)
    return
  }

  res := map[string]interface{}{
    "message": "PTZ Server",
  }

  w.Header().Set("Content-Type", "application/json")

  json.NewEncoder(w).Encode(&res)
}

func handleConnect(w http.ResponseWriter, r *http.Request) {
  if r.Method != "POST" {
    w.WriteHeader(http.StatusNotFound)
    return
  }

  res := map[string]interface{}{
    "code": http.StatusOK,
    "message": "Session started",
    "data": nil,
  }

  var info PTZInfo

  err := json.NewDecoder(r.Body).Decode(&info)

  if err != nil {
    res["code"] = http.StatusBadRequest
    res["message"] = err.Error()
  } else {
    found := false
    sid := ""
    for _, session:= range gSessions {
      if session.ptz.info.Ip == info.Ip && session.ptz.info.Port == info.Port {
        data := SessionID{Id: session.id}
        res["data"] = data
        found = true
        sid = session.id
        fmt.Println("Session exist: " + session.id)
        break
      }
    }

    if !found {
      session, err := NewSession(info.Ip, info.Port, info.Username, info.Password)

      if err != nil {
        res["code"] = http.StatusInternalServerError
        res["message"] = err.Error()
      } else {
        http.SetCookie(w, &http.Cookie{
          Name: "seesion_id",
          Value: session.id,
          Path: "/",
          // MaxAge: 3600,
          // HttpOnly: true,
        })

        gSessions[session.id] = session
      }
    } else {
      http.SetCookie(w, &http.Cookie{
        Name: "seesion_id",
        Value: sid,
        Path: "/",
      })
    }
  }

  w.Header().Set("Content-Type", "application/json")

  json.NewEncoder(w).Encode(&res)
}

func checkCookie(w http.ResponseWriter, r *http.Request) (string, error) {
  cookie, err := r.Cookie("seesion_id")
  if err != nil {
    res := map[string]interface{}{
      "code": http.StatusUnauthorized,
      "message": "Unauthorized",
      "data": nil,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(&res)
    return "", err
  }

  return cookie.Value, nil
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
  if r.Method != "GET" {
    w.WriteHeader(http.StatusNotFound)
    return
  }
  
  sid, err := checkCookie(w, r)

  if err == nil {
    res, err := gSessions[sid].ptz.GetConfigs()

    if err != nil {
      res["code"] = http.StatusInternalServerError
      res["message"] = err.Error()
      res["data"] = nil
    } else {
      gSessions[sid].ActivateSession()
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(&res)
  } else {
    w.WriteHeader(http.StatusUnauthorized)
    return
  }
}

func handlePresets(w http.ResponseWriter, r *http.Request) {
  if r.Method != "GET" {
    w.WriteHeader(http.StatusNotFound)
    return
  }
  
  sid, err := checkCookie(w, r)

  if err == nil {
    res, err := gSessions[sid].ptz.GetPresets()

    if err != nil {
      res["code"] = http.StatusInternalServerError
      res["message"] = err.Error()
      res["data"] = nil
    } else {
      gSessions[sid].ActivateSession()
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(&res)
  } else {
    w.WriteHeader(http.StatusUnauthorized)
    return
  }
}

func handlePosition(w http.ResponseWriter, r *http.Request) {
  if r.Method != "GET" {
    w.WriteHeader(http.StatusNotFound)
    return
  }
  
  sid, err := checkCookie(w, r)

  if err == nil {
    res, err := gSessions[sid].ptz.GetPosition()

    if err != nil {
      res["code"] = http.StatusInternalServerError
      res["message"] = err.Error()
      res["data"] = nil
    } else {
      gSessions[sid].ActivateSession()
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(&res)
  } else {
    w.WriteHeader(http.StatusUnauthorized)
    return
  }
}

func handleMoving(w http.ResponseWriter, r *http.Request) {
  if r.Method != "GET" {
    w.WriteHeader(http.StatusNotFound)
    return
  }
  
  sid, err := checkCookie(w, r)

  if err == nil {
    res, err := gSessions[sid].ptz.IsMoving()

    if err != nil {
      res["code"] = http.StatusInternalServerError
      res["message"] = err.Error()
      res["data"] = nil
    } else {
      gSessions[sid].ActivateSession()
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(&res)
  } else {
    w.WriteHeader(http.StatusUnauthorized)
    return
  }
}

func handleSnapshot(w http.ResponseWriter, r *http.Request) {
  if r.Method != "GET" {
    w.WriteHeader(http.StatusNotFound)
    return
  }

  sid, err := checkCookie(w, r)

  if err == nil {
    res := gSessions[sid].GetSnapshot()
    gSessions[sid].ActivateSession()

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(&res)
  } else {
    w.WriteHeader(http.StatusUnauthorized)
    return
  }
}

func handleProfile(w http.ResponseWriter, r *http.Request) {
  if r.Method != "POST" {
    w.WriteHeader(http.StatusNotFound)
    return
  }

  sid, err := checkCookie(w, r)

  if err == nil {

    res := map[string]interface{}{
      "code": http.StatusBadRequest,
      "message": "Invalid request parameters",
      "data": nil,
    }
  
    var profile Profile
  
    err := json.NewDecoder(r.Body).Decode(&profile)
  
    if err == nil {
      res["code"] = http.StatusOK
      res["message"] = "Change profile"
      res["data"] = profile

      if profile.Name != gSessions[sid].ptz.profile_name {
        gSessions[sid].ActivateSession()
        err = gSessions[sid].ChangeProfile(profile.Name)

        if err != nil {
          res["code"] = http.StatusInternalServerError
          res["message"] = err.Error()
        }
      }
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(&res)
  } else {
    w.WriteHeader(http.StatusUnauthorized)
    return
  }
}

func handleRelativeMove(w http.ResponseWriter, r *http.Request) {
  if r.Method != "POST" {
    w.WriteHeader(http.StatusNotFound)
    return
  }

  sid, err := checkCookie(w, r)

  if err == nil {

    res := map[string]interface{}{
      "code": http.StatusBadRequest,
      "message": "Invalid request parameters",
      "data": nil,
    }
  
    var pos Position
  
    err := json.NewDecoder(r.Body).Decode(&pos)
  
    if err == nil {
      res, _ = gSessions[sid].ptz.MoveRelativePosition(pos.Pan, pos.Tilt, pos.Zoom, pos.PanSpeed, pos.TiltSpeed, pos.ZoomSpeed)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(&res)
  } else {
    w.WriteHeader(http.StatusUnauthorized)
    return
  }
}

func handleGotoPosition(w http.ResponseWriter, r *http.Request) {
  if r.Method != "POST" {
    w.WriteHeader(http.StatusNotFound)
    return
  }

  sid, err := checkCookie(w, r)

  if err == nil {

    res := map[string]interface{}{
      "code": http.StatusBadRequest,
      "message": "Invalid request parameters",
      "data": nil,
    }
  
    var pos Position
  
    err := json.NewDecoder(r.Body).Decode(&pos)
  
    if err == nil {
      res, _ = gSessions[sid].ptz.GotoPosition(pos.Pan, pos.Tilt, pos.Zoom, pos.PanSpeed, pos.TiltSpeed, pos.ZoomSpeed)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(&res)
  } else {
    w.WriteHeader(http.StatusUnauthorized)
    return
  }
}

func handleGotoPreset(w http.ResponseWriter, r *http.Request) {
  if r.Method != "POST" {
    w.WriteHeader(http.StatusNotFound)
    return
  }

  sid, err := checkCookie(w, r)

  if err == nil {

    res := map[string]interface{}{
      "code": http.StatusBadRequest,
      "message": "Invalid request parameters",
      "data": nil,
    }
  
    var preset PTZPresetID
  
    err := json.NewDecoder(r.Body).Decode(&preset)
  
    if err == nil {
      res, _ = gSessions[sid].ptz.GotoPreset(preset.Id)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(&res)
  } else {
    w.WriteHeader(http.StatusUnauthorized)
    return
  }
}

func handleGotoHome(w http.ResponseWriter, r *http.Request) {
  if r.Method != "POST" {
    w.WriteHeader(http.StatusNotFound)
    return
  }

  sid, err := checkCookie(w, r)

  if err == nil {
    res, _ := gSessions[sid].ptz.GotoHome()

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(&res)
  } else {
    w.WriteHeader(http.StatusUnauthorized)
    return
  }
}

func handleStop(w http.ResponseWriter, r *http.Request) {
  if r.Method != "POST" {
    w.WriteHeader(http.StatusNotFound)
    return
  }

  sid, err := checkCookie(w, r)

  if err == nil {
    res, _ := gSessions[sid].ptz.Stop()

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(&res)
  } else {
    w.WriteHeader(http.StatusUnauthorized)
    return
  }
}

type StaticFile struct {
	name string
}

func (file StaticFile) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fileBytes, err := os.ReadFile(file.name)
	if err != nil {
		fmt.Printf("File %s not found.\n", file.name)
    w.WriteHeader(http.StatusNotFound)
    return
	}
  ext := filepath.Ext(file.name)
  mimetype := mime.TypeByExtension(ext)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", mimetype)
  // w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(fileBytes)
}

func check_session_expire() {
  for {
    for id, session:= range gSessions {
      if session.session_end {
        session.image = nil
        session.lock = nil
        session = nil
        fmt.Println("Session expired: " + id)
        delete(gSessions, id)
        break
      }
    }

    time.Sleep(1 * time.Second)
  }
}

func server_main() {

	http.Handle("/", &StaticFile{"static/index.html"})
	http.Handle("/index.html", &StaticFile{"static/index.html"})
	http.Handle("/favicon.ico", &StaticFile{"static/favicon.ico"})
	http.Handle("/js/jquery.js", &StaticFile{"static/js/jquery.js"})
	http.Handle("/js/vue3.js", &StaticFile{"static/js/vue3.js"})
	http.Handle("/js/element-plus.js", &StaticFile{"static/js/element-plus.js"})
	http.Handle("/js/element-plus/icons-vue.js", &StaticFile{"static/js/element-plus/icons-vue.js"})
	http.Handle("/css/element-plus.css", &StaticFile{"static/css/element-plus.css"})

  http.HandleFunc("/ptz", handleApiHome)
  http.HandleFunc("/snapshot", handleSnapshot)
  http.HandleFunc("/ptz/connect", handleConnect)
  http.HandleFunc("/ptz/config", handleConfig)
  http.HandleFunc("/ptz/presets", handlePresets)
  http.HandleFunc("/ptz/position", handlePosition)
  http.HandleFunc("/ptz/moving", handleMoving)
  http.HandleFunc("/ptz/profile", handleProfile)
  http.HandleFunc("/ptz/move/relative", handleRelativeMove)
  http.HandleFunc("/ptz/goto/position", handleGotoPosition)
  http.HandleFunc("/ptz/goto/preset", handleGotoPreset)
  http.HandleFunc("/ptz/goto/home", handleGotoHome)
  http.HandleFunc("/ptz/stop", handleStop)

  fmt.Println("Starting Restful server on port 8000.")

  go check_session_expire()

	err := http.ListenAndServe(":8000", nil)

  if err != nil {
    fmt.Println(err)
  }
}


