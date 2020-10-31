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

type githubInfo struct { //"githubinfo": Is a struct we use to marshal retrieved data into JSON
	Project      string   //Project: The repostory's name
	Owner        string   //Owner: The owner of the repository
	TopCommitter []string //TopCommiter: The top commiter(s) in the repository
	Commits      int      //Commits: The total number of commits to the repository
	Languages    []string //Languages: The programming languges used in the repository
}

type username struct { //"username": Is a struct for unmarshaling user data JSON from api.github.com
	Name    string `json:"name"`    //Name: The organization name of the user
	Login   string `json:"login"`   //Login: The login username of the user
	Type    string `json:"type"`    //Type: What type of user the account is (Organization/User)
	Message string `json:"message"` //Message: The response message api.github.com sends when it could not find something
}

type contributor struct { //"contributor": Is a struct for unmarshaling contributor JSON from api.github.com
	Name          string `json:"login"`         //Name: The login username of the contributor
	Contributions int    `json:"contributions"` //Contribution: The number of contributions the contributor has made
}

func main() {
	addr, err := determineListenAddress()
	if err != nil {
		log.Fatal(err)
	}
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

func serviceHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/") //Get rid of the root path, leaving "github.com/[user]/[repo]"
	path = path[3:]
	ok, message := checkPath(path) //Check the path we have
	if ok {
		info := getInfo(path)         //Create a githubinfo struct using path
		js, err := json.Marshal(info) //Marshal into JSON
		if err != nil {               //Error message
			fmt.Fprintln(w, "Failed to marshal json")
		} else { //Write JSON
			w.Header().Set("Content-Type", "application/json")
			w.Write(js)
		}
	} else { //Error message if path is invalid
		fmt.Fprintln(w, message)
	}
}

/*
*	Checks that the given path conforms to what is expected
*	@param path []string This is a slice containing each of the parts of the request path "github.com/[user]/[repo]"
*	@return bool This is a boolean to determine if the check passed or not
*	@return string This holds an error message if the check did not pass
 */
func checkPath(path []string) (bool, string) {
	ok := false
	message := ""
	if len(path) != 3 {
		message = "Incorrect path length"
	} else if strings.ToLower(path[0]) != "github.com" {
		message = "Not a github link"
	} else {
		ok = true
		message = "All good"
	}
	return ok, message
}

/*
*	Creates a githubinfo object and fills it with the relevant data.
*	If any of the data requests failed, the corresponding object field will be set to "Could not find [thing we didnt find]"
*	@param params []string This is a slice containing each of the parts of the request path "github.com/[user]/[repo]"
*	@return githubinfo This is a githubinfo object containing all the data we wish to marshal as JSON
 */
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

/*
*	Gets the name(s) of the top contributor(s) to the repository
*	@param user string This is the username given in the request path
*	@param param string This is the repository name given in the request path
*	@return []string This is a slice of the top contributer(s) to the repository
*	@return int This is the total number of commits to the repository
*	@return bool This is a boolean tells whether we failed to get the contributors or not
 */
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
				contributors = make([]string, 0, len(data)) //Make contributors become a slice of equal length
				for i := 0; i < len(data); i++ {
					commits += data[i].Contributions  //Update total commits
					if data[i].Contributions == top { //Append to contributors slice if mutliple people share the top
						contributors = append(contributors, data[i].Name)
					} else if data[i].Contributions > top { //Reset slice if a new top is met
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

/*
*	Gets all the programming languges used in the repository
*	@param user string This is the username given in the request path
*	@param param string This is the repository name given in the request path
*	@return []string This is a slice of all the programming languages used in the repository
*	@return bool This is a boolean tells whether we failed to get the languages or not
 */
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
			langs = make([]string, 0, len(data)) //api.github.com returns the languges as keynames rather than values
			for key := range data {              //so we get the keynames and put them into langs
				langs = append(langs, key)
				if key == "message" {
					fail = true
				}
			}
		}
	}
	return langs, fail
}

/*
*	Gets the name of the given user. If it's an organization, we use the organizations name. Otherwise we use the login username
*	@param param string This is the username given in the request path
*	@return string This is the found name of the user. We use the login username if the user is a normal user,
*	and the organization name if it's an organization
*	@return bool This is a boolean tells whether we failed to get the name or not
 */
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
