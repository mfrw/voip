package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
)

// Interfaces

const (
	LOCAL_INTERFACE  = "eth0"
	REMOTE_INTERFACE = "eno1"
	VOIP_SERVER      = "192.168.19.4"
	LOCAL_HOST       = "192.168.2.81"
)

// ROUTERS and USERS

const (
	RS1        = "192.168.2.83"
	RS2        = "192.168.2.84"
	RS3        = "192.168.2.85"
	RELAY_USER = "mfrw"
)

// CALLING info

const (
	MAX_CALL               = 300
	PARALLEL_CALLS         = 200
	VOIP_SERVER_PORT       = 5080
	REMOTE_HOST_SSH        = VOIP_SERVER
	REMOTE_USER            = "root"
	REMOTE_HOST_IPERF_PORT = 2222
)

var WORKING_DIR, _ = os.Getwd()

// SIP details

const (
	SIP_CALLER_ID          = 1000
	SIP_CALLER_PASSWORD    = "killerbee#121"
	SIP_DESTINATION_NUMBER = 999999
	CALL_TIMEOUT           = 35
	CALL_SAMPLE_RATE       = 8000
)

func isRoot() bool {
	return os.Geteuid() == 0
}

func startCollectl() {
	defer log.Println("Done with starting collectl")

	// prepare the commands
	cmd1 := exec.Command("sh", "-c", "sudo -u mfrw collectl --all -oT -f collectl.output")
	cmd2 := exec.Command("sh", "-c", "sudo -u mfrw ssh mfrw@192.168.2.83 collectl --all -oT -f collectl.output")
	cmd3 := exec.Command("sh", "-c", "sudo -u mfrw ssh mfrw@192.168.2.84 collectl --all -oT -f collectl.output")
	cmd4 := exec.Command("sh", "-c", "sudo -u mfrw ssh mfrw@192.168.2.85 collectl --all -oT -f collectl.output")
	cmd5 := exec.Command("sh", "-c", "sudo -u mfrw ssh falak@192.168.2.13 collectl --all -oT -f collectl.output")

	// Now run the commands and dont wait for them .. background
	err := cmd1.Start()
	if err != nil {
		log.Println("could not run collectl on host")
	}
	err = cmd2.Start()
	if err != nil {
		log.Println("could not run collectl on 83")
	}
	err = cmd3.Start()
	if err != nil {
		log.Println("could not run collectl on 84")
	}
	err = cmd4.Start()
	if err != nil {
		log.Println("could not run collectl on 85")
	}
	err = cmd5.Start()
	if err != nil {
		log.Println("could not run collectl on 13")
	}
}

// calculate b/w using iperf

func iperf(callNr int) {
	log.Println("iperf BW calculation started")
	defer log.Println("*****iperf ended*****")
	iperfArgs := fmt.Sprintf("iperf -c %s -p %s -i 10 -f -K -P %d >> iperf_%d.txt", VOIP_SERVER, REMOTE_HOST_IPERF_PORT, callNr, callNr)
	cmd := exec.Command("sh", "-c", iperfArgs)
	cmd.Run()

}

// Calculate RTT using fping
func fping(callNr int) {
	log.Println("fping RTT measurement started")
	defer log.Println("**********fping ended*********")
	fpingArgs := fmt.Sprintf("sudo -u mfrw fping -i 1000 -c 20 -eD %s > fping_%d.txt", VOIP_SERVER, callNr)
	cmd := exec.Command("sh", "-c", fpingArgs)
	cmd.Run()
}

// Kill preexisting (TODO:its very crude so make it a little sane)

func zapExisting() bool {
	return true
}

func call(calls int) {
	var wg sync.WaitGroup // for syncing parallel calls

	for i := 1; i <= calls; i++ {
		wg.Add(1)

		go func(nrCall, TotalCalls int) {
			log.Println("Started call", nrCall, "of", TotalCalls)
			tmpFile := fmt.Sprintf("REC_%d_%d", TotalCalls, nrCall)
			//localtime := time.Now().UnixNano()
			sound_file := tmpFile
			//pkt_dmp := tmpFile + ".pcap"
			defer wg.Done()
			SIP_caller_id := SIP_CALLER_ID + (nrCall % 20)
			SIP_dst_num := SIP_DESTINATION_NUMBER - nrCall
			cwd, _ := os.Getwd()
			callerArgs := fmt.Sprintf("docker run --rm -v %s:/data -w /data mfrw/pjsua-voip python ./../record-samples.py %s %s %d %d %s %d %d %d", cwd, sound_file, VOIP_SERVER, VOIP_SERVER_PORT, SIP_caller_id, SIP_CALLER_PASSWORD, SIP_dst_num, CALL_TIMEOUT, CALL_SAMPLE_RATE)
			log.Println(callerArgs)
			cmd := exec.Command("sh", "-c", callerArgs)
			//otpt, err := cmd.CombinedOutput()
			//log.Println(string(otpt))
			err := cmd.Run()
			if err != nil {
				log.Println(err)
			}

		}(i, calls)

	}
	wg.Wait()
	log.Printf("Recording done for %d parallel Calls\n", calls)
}

// call parallely using goroutines (glorified threads)
func parallel_calls() {
	for i := 32; i <= PARALLEL_CALLS; i++ {
		iperf(i)

		//		fping(i)

		// make a call
		call(i)

	}
}

func main() {
	// check for root
	if !isRoot() {
		log.Fatalf("You need to be root to run this")
	}

	// Start the collectl process in background
	// startCollectl()

	parallel_calls()

}
