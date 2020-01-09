package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/user"
	"runtime"
	"time"
)

// Info about current machine
type Info struct {
	IP      string `json:"ip"`
	Name    string `json:"name"`
	User    string `json:"user"`
	Version string `json:"version"`
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

func getUser() (string, error) {
	user, err := user.Current()
	return user.Username, err
}

func getVersion() string {
	version := os.Getenv("VERSION")
	if len(version) == 0 {
		return "0"
	}
	return version
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
	db := Storage{}
	// err := db.Connect("postgres", "1", "localhost", "5432", "gotest", "disable")
	err := db.Init()
	if err != nil {
		log.Fatal(err)
	}
	err = createUserTable(db.DB)
	if err != nil {
		log.Fatal(err)
	}

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

		userName, err := getUser()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		write(w, Info{
			IP:      ip,
			Name:    name,
			User:    userName,
			Version: getVersion(),
		})
	})
	http.HandleFunc("/crash", func(w http.ResponseWriter, r *http.Request) {
		write(w, "exited")
		log.Fatal("crash")
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
	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		write(w, "test")
	})
	http.HandleFunc("/db", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.DB.Query("SELECT id, name, email FROM users")
		if err != nil {
			log.Fatal(err)
		}

		type user struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		users := []user{}
		for rows.Next() {
			var id int
			var name, email string
			rows.Scan(&id, &name, &email)
			users = append(users, user{
				ID:    id,
				Name:  name,
				Email: email,
			})
		}
		write(w, users)
	})

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "80"
	}
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func createUserTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name character varying(256),
		email character varying(256)
	);`)
	if err != nil {
		log.Fatal(err)
		return err
	}
	_, err = db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS users_pkey ON users(id int4_ops);")
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}
