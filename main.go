package main

import (
	"encoding/json"
	"errors"
	"flag"
	irsdk "github.com/quimcalpe/iracing-sdk"
	"gopkg.in/ini.v1"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var sdk irsdk.IRSDK
var homeTemplate *template.Template
var vehiclesCfg map[string][]int
var self struct {
	isInit bool
	driver irsdk.Driver
}

func main() {
	err := initConf()
	if err != nil {
		log.Fatal(err)
	}

	sdk = irsdk.Init(nil)
	defer sdk.Close()

	h, err := template.ParseFiles("index.html")
	if err != nil {
		log.Fatal(err)
	}
	homeTemplate = h

	flag.Parse()
	log.SetFlags(0)
	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/", home)
	log.Printf("Listening on %q", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

var addr = flag.String("addr", "localhost:8080", "http service address")

var upgrader = websocket.Upgrader{}

type data struct {
	IsConnected    bool
	RPMLights      rpmLights
	EngineWarnings interface{}
	DisplayUnits   interface{}
	Gear           interface{}
	RPM            interface{}
	AbsActive      interface{}
}

type rpmLights struct {
	First    float64
	Last     float64
	Blink    float64
	Shift    float64
	HasGears bool
	Gears    []int
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer func(c *websocket.Conn) {
		err := c.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(c)

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			break
		}
		online := true
		for {
			sdk.WaitForData(17 * time.Millisecond)

			rpmL, err := getRPMData(&sdk)
			checkErr(err)

			engineWarnings, err := sdk.GetVar("EngineWarnings")
			checkErr(err)
			units, err := sdk.GetVar("DisplayUnits")
			checkErr(err)
			gear, err := sdk.GetVar("Gear")
			checkErr(err)
			rpm, err := sdk.GetVar("RPM")
			checkErr(err)
			absActive, err := sdk.GetVar("BrakeABSactive")
			checkErr(err)

			d := data{
				sdk.IsConnected(),
				rpmL,
				engineWarnings.Value,
				units.Value,
				gear.Value,
				rpm.Value,
				absActive.Value,
			}

			message, err = json.Marshal(d)
			if err != nil {
				log.Println("error json: ", err)
				break
			}
			err = c.WriteMessage(mt, message)
			if err != nil {
				//log.Println("error write: ", err)
				break
			}
			if sdk.IsConnected() {
				time.Sleep(17 * time.Millisecond)
				if !online {
					log.Println("iRacing connected!")
				}
				online = true
			} else {
				time.Sleep(3 * time.Second)
				if online {
					log.Println("Waiting for iRacing connection...")
				}
				online = false
			}
		}
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	err := homeTemplate.Execute(w, "ws://"+r.Host+"/ws")
	if err != nil {
		log.Fatal(err)
	}
}

func checkErr(err error) {
	if err != nil {
		//log.Println(err)
	}
}

func getRPMData(sdk *irsdk.IRSDK) (rpmLights, error) {
	first := sdk.GetSession().DriverInfo.DriverCarSLFirstRPM
	last := sdk.GetSession().DriverInfo.DriverCarSLLastRPM
	blink := sdk.GetSession().DriverInfo.DriverCarSLBlinkRPM
	shift := sdk.GetSession().DriverInfo.DriverCarSLShiftRPM

	hasGears := true
	driver, err := getSelf(sdk)
	checkErr(err)
	if err != nil {
		hasGears = false
	}

	gears, err := getRpmByCar(driver.CarPath)
	checkErr(err)
	if err != nil {
		hasGears = false
	}
	return rpmLights{
		First:    first,
		Last:     last,
		Blink:    blink,
		Shift:    shift,
		HasGears: hasGears,
		Gears:    gears,
	}, nil
}

func getSelf(sdk *irsdk.IRSDK) (irsdk.Driver, error) {
	if self.isInit && sdk.GetSession().DriverInfo.DriverCarIdx == self.driver.CarIdx {
		return self.driver, nil
	}

	for _, driver := range sdk.GetSession().DriverInfo.Drivers {
		if sdk.GetSession().DriverInfo.DriverCarIdx == driver.CarIdx {
			self.isInit = true
			self.driver = driver

			return self.driver, nil
		}
	}

	return irsdk.Driver{}, errors.New("driver not found")
}

func initConf() error {
	vehiclesCfg = map[string][]int{}
	file, err := ioutil.ReadFile("./config/cfg.ini")
	if err != nil {
		return err
	}
	cfg, err := ini.Load(file)
	if err != nil {
		return err
	}
	for _, key := range cfg.Section("").Keys() {
		for _, v := range strings.Split(key.Value(), ",") {
			t, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			vehiclesCfg[key.Name()] = append(vehiclesCfg[key.Name()], t)
		}
	}

	return nil
}

func getRpmByCar(car string) ([]int, error) {
	rpm, ok := vehiclesCfg[car]
	if ok == false {
		return []int{}, errors.New("config not found")
	}
	return rpm, nil
}
