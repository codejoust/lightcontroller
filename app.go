package main


import (
	"github.com/tarm/goserial"
	"net/http"
	"flag"
	"strconv"
	"log"
	"io"
	"fmt"
	"encoding/json"
	"time"
)

var serialPort = flag.String("serial", "/dev/master", "Serial Port Path to Use")
var webPort = flag.Int("port", 8080, "Webserver port to use")

type Light struct {
	Id int
	Val int
	Name string
}

type RoomInfo struct {
	motion_on_time time.Time
	motion_on bool
	door_opened_time time.Time
	door_opened bool
}

var allLights = []Light{
	Light{1, 0, "Main Room Left"},
	Light{2, 0, "Main Room Right"},
	Light{3, 0, "Left front"},
	Light{4, 0, "Right front"},
	Light{5, 0, "led green"},
	Light{6, 0, "led red"},
	Light{7, 0, "led blue"},
}

var SerialPort io.ReadWriteCloser

func connectSerial() {
	c := &serial.Config{Name: *serialPort, Baud: 19200}
    s, err := serial.OpenPort(c)
    if err != nil {
    	log.Fatal(err)
    }
    SerialPort = s
}

func fadeLight(lightchg Light, value int){
	lightDelta := 1
	if lightchg.Val > value { lightDelta = -1 }
	for lightchg.Val != value {
		lightchg.Val += lightDelta
		time.Sleep(time.Millisecond * 10)
		changeVal(lightchg)
	}

}

func changeVal(light Light){
	n, err := fmt.Fprintf(SerialPort, "%dc%dw\n", light.Id, light.Val)
	//log.Println("%dc%dw\n", light.id, light.val)
	if err != nil {
		log.Fatal(err)
		log.Fatal(n)
	}
}

func homePageHandler(w http.ResponseWriter, r *http.Request){
	http.ServeFile(w, r, "views/index.html")
}

func getLightInfoHandler(w http.ResponseWriter, r *http.Request){
	sendLightInfo(w)
}

func sendLightInfo(w http.ResponseWriter){
	enc := json.NewEncoder(w)
	if err := enc.Encode(&allLights); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func changeLightHandler(w http.ResponseWriter, r *http.Request){
	//lightIdInput := r.URL.Path[len("/light/"):]
	//lightId, err := strconv.Atoi(lightIdInput)
	lightId, err := strconv.Atoi(r.FormValue("light"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	value, err := strconv.Atoi(r.FormValue("value"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for num, el := range allLights {
		if el.Id == lightId {
			fmt.Fprintf(w, "/* Changing light, %s to %d */\n", el.Name, value)
			go fadeLight(el, value)
			allLights[num].Val = value
			sendLightInfo(w)
			return
		}
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func turnOffAllLightsHandler(w http.ResponseWriter, r *http.Request){
	for num, el := range allLights {
		if el.Val > 0 {
			go fadeLight(el, 0)
			allLights[num].Val = 0
		}
	}
	sendLightInfo(w)
}

func main(){
	flag.Parse();

	connectSerial()
	
	http.HandleFunc("/", homePageHandler)
	http.HandleFunc("/lights", getLightInfoHandler)
	http.HandleFunc("/light/", changeLightHandler)
	http.HandleFunc("/blackout", turnOffAllLightsHandler)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *webPort), nil))

}




