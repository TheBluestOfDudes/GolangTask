package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const failtext = "Could not find"

type githubInfo struct {
	Project      string
	Owner        string
	TopCommitter []string
	Commits      int
	Languages    []string
}

type username struct {
	Name    string `json:"name"`
	Login   string `json:"login"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

type contributor struct {
	Name          string `json:"login"`
	Contributions int    `json:"contributions"`
}

func main() {
	addr, err := determineListenAddress()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", hello)
	http.HandleFunc("/projectinfo/v1/", serviceHandler)
	log.Printf("Listening on %s...\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}

func determineListenAddress() (string, error) {
	port := os.Getenv("PORT")
	if port == "" {
		return "", fmt.Errorf("$PORT not set")
	}
	return ":" + port, nil
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hi")
}

//Handles the requests to /projectinfo/v1/
func serviceHandler(w http.ResponseWriter, r *http.Request) {
	//Get the URI from browser path
	path := strings.Split(r.URL.Path, "/")
	path = path[3:]
	//Make sure it's ok
	ok, message := checkPath(path)
	if ok {
		info := getInfo(path)
		js, err := json.Marshal(info)
		if err != nil {
			fmt.Fprintln(w, "Failed to marshal json")
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Write(js)
		}
	} else {
		fmt.Fprintln(w, message)
	}
}

//Checks if the given path is of the right legnth and starts with github.
func checkPath(path []string) (bool, string) {
	ok := false
	message := ""
	//Length should be 3. "[github.com,[org/username],[repo]]"
	if len(path) != 3 {
		message = "Incorrect path length"
	} else if strings.ToLower(path[0]) != "github.com" { //The path schould start with "github.com"
		message = "Not a github link"
	} else { //Assuming the above two are fine, the path is deemed useable
		ok = true
		message = "All good"
	}
	return ok, message
}

func getInfo(params []string) githubInfo {
	gI := githubInfo{}
	gI.Project = params[2]
	uName, err := getName(params[1])
	if !err {
		gI.Owner = uName
	} else {
		gI.Owner = failtext + " owner name"
	}
	languages, err := getLanguages(params[1], params[2])
	if !err {
		gI.Languages = languages
	} else {
		gI.Languages = []string{failtext + " languages"}
	}
	contributors, commits, err := getContributor(params[1], params[2])
	if !err {
		gI.TopCommitter = contributors
		gI.Commits = commits
	} else {
		gI.TopCommitter = []string{failtext + " contributors"}
		gI.Commits = 0
	}
	return gI
}

func getContributor(user string, param string) ([]string, int, bool) {
	var contributors []string
	var data []contributor
	fail := false
	commits := 0
	apiLink := "https://api.github.com/repos/" + user + "/" + param + "/contributors"
	resp, err := http.Get(apiLink)
	if err != nil {
		resp.Body.Close()
		fail = true
	} else {
		if resp.Body != nil {
			defer resp.Body.Close()
		} else {
			resp.Body.Close()
			fail = true
		}
		body, readErr := ioutil.ReadAll(resp.Body)
		if readErr != nil {
			fail = true
		} else {

			jsonErr := json.Unmarshal(body, &data)
			if jsonErr != nil {
				fail = true
			} else {
				top := 0
				contributors = make([]string, 0, len(data))
				for i := 0; i < len(data); i++ {
					commits += data[i].Contributions
					if data[i].Contributions == top {
						contributors = append(contributors, data[i].Name)
					} else if data[i].Contributions > top {
						contributors = make([]string, 0, len(data))
						contributors = append(contributors, data[i].Name)
						top = data[i].Contributions
					}
				}
			}
		}
	}
	return contributors, commits, fail
}

func getLanguages(user string, param string) ([]string, bool) {
	var langs []string
	var data map[string]interface{}
	fail := false
	apiLink := "https://api.github.com/repos/" + user + "/" + param + "/languages"
	resp, err := http.Get(apiLink)
	if err != nil {
		resp.Body.Close()
		fail = true
	} else {
		if resp.Body != nil {
			defer resp.Body.Close()
		} else {
			resp.Body.Close()
			fail = true
		}
		body, readErr := ioutil.ReadAll(resp.Body)
		if readErr != nil {
			fail = true
		} else {
			jsonErr := json.Unmarshal(body, &data)
			if jsonErr != nil {
				fail = true
			} else if len(data) == 0 {
				fail = true
			}
			langs = make([]string, 0, len(data))
			for key := range data {
				langs = append(langs, key)
				if key == "message" {
					fail = true
				}
			}
		}
	}
	return langs, fail
}

func getName(param string) (string, bool) {
	var uName username
	name := ""
	apiLink := "https://api.github.com/users/" + param
	fail := false

	resp, err := http.Get(apiLink)

	if err != nil { //Close response if there is an error
		fail = true
		resp.Body.Close()
	} else {
		if resp.Body != nil { //Close response once we're done
			defer resp.Body.Close()
		} else { //Close if the body's empty
			resp.Body.Close()
			fail = true
		}
		body, readErr := ioutil.ReadAll(resp.Body) //Read the body
		if readErr != nil {
			fail = true
		} else { //Get the name from the response body
			uName = username{}
			jsonErr := json.Unmarshal(body, &uName)
			if jsonErr != nil {
				fail = true
			} else if strings.ToLower(uName.Message) == "not found" {
				fail = true
			}
			if strings.ToLower(uName.Type) == "organization" {
				name = uName.Name
			} else {
				name = uName.Login
			}
		}
	}
	return name, fail
}
