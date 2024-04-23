package main

import (
	"encoding/json"
	"fmt"
	"time"
)

func printJson(obj interface{}) {
	json, err := json.Marshal(obj)

	if err != nil {
    fmt.Println("转换为json错误")
  }

	fmt.Println(string(json))
}

func waitForStop(ptz *PTZControl, timeout int) {
	fmt.Println("Wait for stop")
	i := 0
	for i = 0; i <= timeout; i++ {
		time.Sleep(500 * time.Millisecond)
		res, _ := ptz.IsMoving()
		status := res["data"]
		printJson(status)
		if !status.(Moving).Moving {
			break
		}
	}

	if (i == timeout) {
		fmt.Println("Timeout")
	} else {
		fmt.Println("Stopped")
	}
}

func test() {
	info, err := LoadConfig()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	ptz, err := NewPTZControl(info.Ip, info.Port, info.Username, info.Password)
	
	if err != nil {
		fmt.Println("Failed to connect to  IP camera.")
		return
	}

	printJson(ptz.configs)
	printJson(ptz.profiles)
	fmt.Println("Current Profile: " + ptz.profile_name)

	for key := range ptz.profiles {
		ptz.SetProfile(key)
		fmt.Println("Set Profile: " + ptz.profile_name)
		res, _ := ptz.GetStreamUri()
		printJson(res["data"])
	}

	res, _ := ptz.IsMoving()
	printJson(res["data"])
	res, _ = ptz.GetPosition()
	printJson(res["data"])

	res, _ = ptz.GetConfigs()
	printJson(res["data"])

	res, _ = ptz.GetPresets()
	printJson(res["data"])

	fmt.Println("Goto Preset 1")
	res, _ = ptz.GotoPreset("1")
	printJson(res["data"])

	waitForStop(ptz, 10)
	time.Sleep(2 * time.Second)

	_, _ = ptz.GotoPosition(0.8, 0.8, 0, 0.5, 0.5, 0.5)

	time.Sleep(1 * time.Second)
	fmt.Println("Send Stop")
	res, _ = ptz.Stop()
	printJson(res)


	waitForStop(ptz, 10)
	time.Sleep(2 * time.Second)

	fmt.Println("Goto Home")
	res, _ = ptz.GotoHome()
	printJson(res)

	waitForStop(ptz, 10)
	time.Sleep(2 * time.Second)

	fmt.Println("move to 0.2, -0.2, 0.2")
	res, _ = ptz.GotoPosition(0.2, -0.2, 0, 1, 1, 1)
	printJson(res)

	waitForStop(ptz, 10)
	time.Sleep(2 * time.Second)

	fmt.Println("relative move -0.1, 0.1")
	res, _ = ptz.MoveRelativePosition(-0.1, 0.1, 0, 1, 1, 1)
	printJson(res)

	waitForStop(ptz, 10)
	time.Sleep(2 * time.Second)

	fmt.Println("Goto Preset 2")
	res, _ = ptz.GotoPreset("2")
	printJson(res["data"])

	waitForStop(ptz, 10)
}


func main() {
	// test()
	// server_main()
	server_gin_main()
}