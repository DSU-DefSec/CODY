package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
    "encoding/json"
)

func makeRequest(req *http.Request, data url.Values) (string, int, error) {
	client := http.Client{}
    // can i not hardcode this lol
    req.SetBasicAuth("ghost_of_cutshaw", "Password1!")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	resp, err := client.Do(req)
	if err != nil {
		return "", 400, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		err = errors.New(string(body))
	}
    return string(body), resp.StatusCode, err
}

func vcloudAuth(username string, password string) error {
	data := url.Values{}
	data.Set("username", username)
	data.Add("password", password)
	req, err := http.NewRequest("POST", webDeployAPI+"/auth", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
    _, _, err = makeRequest(req, data)
	return err
}

func vappDeployUser(vapp string, destination string) (string, error) {

    if ! validateName(vapp) {
        return "", errors.New("Invalid name")
    }

	data := url.Values{}
	data.Set("type", "user")
	data.Add("scheduled", "false")
	data.Add("vapp", vapp)
	data.Add("destination", destination)
	req, err := http.NewRequest("POST", webDeployAPI+"/deploy", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
    body, status, err := makeRequest(req, data)

	if status >= 400 {
		if status == 406 {
			return "", errors.New("Invalid name")
		} else if status == 409 {
            id := vapp{}
            err := json.Unmarshal([]byte(body), &id)
            if err != nil {
    			return "", err
            } else {
    			return id.Id, errors.New("Already deployed")
            }
		} else {
			return "", errors.New(body)
		}
	}
	return body, err
}