package main

// Licensing: Apache-2.0 and BSD-3-Clause
/*
 *  Copyright (c) 2017 Wind River Systems, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at:
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software  distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES
 * OR CONDITIONS OF ANY KIND, either express or implied.
 */

// License BSD-3-Clause https://github.com/google/go-github/blob/master/LICENSE
// for github.com/google/go-github/github

// Supporting documentation:
// 	https://godoc.org/github.com/google/go-github/github
// 	https://developer.github.com/v3/repos/#create
// 	https://developer.github.com/v3/repos/#delete

import (
	//"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	//"reflect" //fmt.Println("Type:", reflect.TypeOf(f))
	"strconv"

	"github.com/google/go-github/github"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
)

var MAIN_config Configuration // Server configuration structure

const (
	config_file = "./g-snoop_config.json"
)

type Configuration struct {
	GithubAccount       string `json:"account"`
	GithubURL           string `json:"github_url"`
	GithubToken         string `json:"token"`
	HttpPort            int    `json:"http_port"`
	Debug_On            bool   `json:"debug_on"`
	Verbose_On          bool   `json:"verbose_on"`
	ConfigReloadAllowed bool   `json:"config_reload_allowed"`
}

func GetConfigurationInfo(configuration *Configuration, first_time bool) {

	// When this func is called the first time we want to load the config file
	// regardless of whether the value of config_reload_allowed is true.
	// If this config variable is false on furture invocations we do not
	// want to allow the config file to be reloaded. We created a temp struct
	// to load and check config_reload_allowed first. If this varible is true
	// then we can proceed to load the current values.

	var temp_config Configuration

	file, _ := os.Open(config_file)
	decoder := json.NewDecoder(file)

	temp_config = Configuration{}
	//*configuration = Configuration{}
	err := decoder.Decode(&temp_config)
	if err != nil {
		fmt.Println("error:", err)
	}
	if first_time || temp_config.ConfigReloadAllowed {
		*configuration = temp_config
		if MAIN_config.Verbose_On {
			fmt.Println("Configuration:")
			fmt.Println("-----------------------------------------------")
			fmt.Println("github account		= ", configuration.GithubAccount)
			fmt.Println("github url	= ", configuration.GithubURL)
			fmt.Println("token 		= ", configuration.GithubToken)
			fmt.Println("http port	= ", configuration.HttpPort)
			fmt.Println("debug on	  = ", configuration.Debug_On)
			fmt.Println("verbose on	  =", configuration.Verbose_On)
			fmt.Println("config  reload		= ", configuration.ConfigReloadAllowed)
		}
	}
}

var index_html_1of3 *bytes.Buffer
var index_html_3of3 *bytes.Buffer

// func CreateHomePage () *bytes.Buffer {
func CreateHomePage() string {
	// get last repos list
	// write to file f2
	// concat f1, f2, f3

	// list public repositories for org "Wind River"
	//opt := &github.RepositoryListByOrgOptions{Type: "public"}
	//repos, _, err := client.Repositories.ListByOrg(context, "wind-river", opt)
	// delete an existing repository
	// Documentation: https://godoc.org/github.com/google/go-github/github
	context := context.Background()
	github_client := github.NewClient(nil)
	repos, _, err := github_client.Repositories.List(context, MAIN_config.GithubAccount, nil)
	if err != nil {
		fmt.Println("Repositories.List(...) returned error: %v", err)
		// TODO: Need to handle this error more gracefully
	}

	if err != nil || len(repos) == 0 {
		fmt.Println("Error!")
		return ""
	}

	// https://godoc.org/github.com/google/go-github/github#Repository
	type RepoRecord struct {
		ID            int
		Name          string
		Description   string
		Language      string
		StarsCount    int
		ForksCount    int
		Size          int
		WatchersCount int
	}

	// Iterate through the list of repos and create html for home page
	var new_html string
	for i := 0; i < len(repos); i++ {
		var description string
		var language string
		// crashes when value string points to nil.
		if repos[i].Description == nil {
			description = ""
		} else {
			description = *repos[i].Description
		}
		if repos[i].Language == nil {
			language = ""
		} else {
			language = *repos[i].Language
		}
		record := &RepoRecord{
			ID:            *repos[i].ID,
			Name:          *repos[i].Name,
			Description:   description,
			Language:      language,
			StarsCount:    *repos[i].StargazersCount,
			ForksCount:    *repos[i].ForksCount,
			Size:          *repos[i].Size,
			WatchersCount: *repos[i].WatchersCount}

		// add another row in the repo table listing
		new_html += fmt.Sprintf("<tr> \n  <td> %s </td>  \n  <td> %s </td>  \n  <td> %d </td> \n  <td> %d </td> \n  <td> %s </td>\n </tr> \n",
			record.Name, record.Language, record.StarsCount,  record.Size, record.Description)
		fmt.Println (i)
	}
	fmt.Println (new_html)
	// Splice together the three templates to dynamically create a new homepage
	return fmt.Sprintf("%s %s %s", index_html_1of3.String(), new_html, index_html_3of3.String())
}

// Print useful info
func displayURLRequest(request *http.Request) {
	fmt.Println()
	fmt.Println("-----------------------------------------------")
	fmt.Println("URL Request: ", request.URL.Path)
	log.Println()
	fmt.Println("query params were:", request.URL.Query())
}

func MgtAccoutnHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("in MgtAccoutHandler")
	fmt.Println("GET params were:", r.URL.Query())
	// if only one expected
	repo_name := r.URL.Query().Get("repo_name")
	user_request := r.URL.Query().Get("radio_button")
	fmt.Println(repo_name, user_request)
	if repo_name != "" {
		// We have a repo name. We need to take action. First we need to
		// set up the authenticate token with the github server.
		context := context.Background()
		// get token from: https://github.com/settings/tokens/new
		// and you need to enter it in the configuration file
		tokenService := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: MAIN_config.GithubToken})
		tokenClient := oauth2.NewClient(context, tokenService)
		github_client := github.NewClient(tokenClient)

		if user_request == "create" {
			// create a new public repository
			repo := &github.Repository{
				Name:    github.String(repo_name),
				Private: github.Bool(false),
			}
			github_client.Repositories.Create(context, "", repo)

		} else if user_request == "delete" {
			// delete an existing repository
			github_client.Repositories.Delete(context, MAIN_config.GithubAccount, repo_name)

		}
	}
	// regenerate the home page in both cases of success or error
	w.Header().Set("Content-Type", "text/html")
	io.WriteString(w, CreateHomePage())
}

// HomeHandler will be rendering of the home page
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("in HomeHandler")
	w.Header().Set("Content-Type", "text/html")
	io.WriteString(w, CreateHomePage())
}

// Pretty print (format) the json reply.
func httpSendReply(http_reply http.ResponseWriter, data interface{}) {

	// We want to pretty print the json reply. We need to wrap:
	//    json.NewEncoder(http_reply).Encode(reply)
	// with the following code:

	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent("", "   ") // tells how much to indent "  " spaces.
	err := encoder.Encode(data)

	if MAIN_config.Debug_On {
		displayURLReply(buffer.String())
	}

	if err != nil {
		io.WriteString(http_reply, "error - could not encode reply")
	} else {
		io.WriteString(http_reply, buffer.String())
	}
}

// Display debug info about a url request
func displayURLReply(url_reply string) {
	// Display http reply content for monitoring and testing purposes
	fmt.Println("-----------------------------------------------")
	fmt.Println("URL Reply:")
	fmt.Println("---------------:")
	fmt.Println(url_reply)
}

// We provide a response when an api call is made but not found
func notFound(http_reply http.ResponseWriter, request *http.Request) {
	if MAIN_config.Debug_On {
		displayURLRequest(request)
	} // display url data

	type notFoundReply struct {
		Status           string `json:"status"`
		Message          string `json:"error_message"`
		DocumentationUrl string `json:"documentation_url"`
	}

	replyData := notFoundReply{Status: "failed",
		Message:          "Not Found",
		DocumentationUrl: "TBD"}
	httpSendReply(http_reply, replyData)
}

func main() {
	// Get server configuration info
	GetConfigurationInfo(&MAIN_config, true)

	// Load home page templates 1-of-3 & 3-of-3. We will auto generate 2-of-3 when served.
	// Read first of three templates
	index_html_1of3 = bytes.NewBuffer(nil)
	f1, err := os.Open("index1of3.html")
	if err != nil {
		log.Fatal(err)
	}
	io.Copy(index_html_1of3, f1)
	f1.Close()

	// Read 3rd of three templates
	index_html_3of3 = bytes.NewBuffer(nil)
	f2, err := os.Open("index3of3.html")
	if err != nil {
		log.Fatal(err)
	}
	io.Copy(index_html_3of3, f2)
	f2.Close()

	r := mux.NewRouter()
	r.HandleFunc("/api/snoop/mgt_account", MgtAccoutnHandler).Methods("GET")
	r.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	r.HandleFunc("/", HomeHandler).Methods("GET")
	r.NotFoundHandler = http.HandlerFunc(notFound)
	fmt.Println("running server on:", MAIN_config.HttpPort)
	// Create port string, e.g., for port 8080 we create ":8080" needed for ListenAndServe ()
	port_str := ":" + strconv.Itoa(MAIN_config.HttpPort)
	http.ListenAndServe(port_str, r)
}
