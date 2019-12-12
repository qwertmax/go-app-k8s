package main

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"
)

// Info about current machine
type Info struct {
	IP   string `json:"ip"`
	Name string `json:"name"`
}

func getIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		// fmt.Printf("%v\n", addrs[0])
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				return v.IP.String(), nil
			case *net.IPAddr:
				return v.IP.String(), nil
			}
		}
	}

	return "", errors.New("unable to get IP")
}

func getHostname() (string, error) {
	return os.Hostname()
}

func write(w http.ResponseWriter, body interface{}) error {
	js, err := json.Marshal(body)
	if err != nil {
		return err
	}

	w.Write(js)
	w.Header().Set("Content-Type", "application/json")
	return nil
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ip, err := getIP()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		name, err := getHostname()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		write(w, Info{
			IP:   ip,
			Name: name,
		})
	})
	http.HandleFunc("/crash", func(w http.ResponseWriter, r *http.Request) {
		os.Exit(3)
	})

	http.HandleFunc("/load", func(w http.ResponseWriter, r *http.Request) {
		done := make(chan int)

		for i := 0; i < runtime.NumCPU(); i++ {
			go func() {
				for {
					select {
					case <-done:
						return
					default:
					}
				}
			}()
		}

		time.Sleep(time.Second * 10)
		close(done)

		write(w, "done")
	})

	http.ListenAndServe(":80", nil)
}
