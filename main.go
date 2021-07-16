package main

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type Robot struct {
	XMLName   xml.Name `xml:"robot"`
	Text      string   `xml:",chardata"`
	Generator string   `xml:"generator,attr"`
	Generated string   `xml:"generated,attr"`
	Rpa       string   `xml:"rpa,attr"`
	Suite     struct {
		Text   string `xml:",chardata"`
		ID     string `xml:"id,attr"`
		Name   string `xml:"name,attr"`
		Source string `xml:"source,attr"`
		Suite  struct {
			Text   string `xml:",chardata"`
			ID     string `xml:"id,attr"`
			Name   string `xml:"name,attr"`
			Source string `xml:"source,attr"`
			Test   []struct {
				Text string `xml:",chardata"`
				ID   string `xml:"id,attr"`
				Name string `xml:"name,attr"`
				Kw   []struct {
					Text      string `xml:",chardata"`
					Name      string `xml:"name,attr"`
					Library   string `xml:"library,attr"`
					Type      string `xml:"type,attr"`
					Doc       string `xml:"doc"`
					Arguments struct {
						Text string   `xml:",chardata"`
						Arg  []string `xml:"arg"`
					} `xml:"arguments"`
					Msg struct {
						Text      string `xml:",chardata"`
						Timestamp string `xml:"timestamp,attr"`
						Level     string `xml:"level,attr"`
					} `xml:"msg"`
					Status struct {
						Text      string `xml:",chardata"`
						Status    string `xml:"status,attr"`
						Starttime string `xml:"starttime,attr"`
						Endtime   string `xml:"endtime,attr"`
					} `xml:"status"`
				} `xml:"kw"`
				Status struct {
					Text      string `xml:",chardata"`
					Status    string `xml:"status,attr"`
					Starttime string `xml:"starttime,attr"`
					Endtime   string `xml:"endtime,attr"`
					Critical  string `xml:"critical,attr"`
				} `xml:"status"`
			} `xml:"test"`
			Status struct {
				Text      string `xml:",chardata"`
				Status    string `xml:"status,attr"`
				Starttime string `xml:"starttime,attr"`
				Endtime   string `xml:"endtime,attr"`
			} `xml:"status"`
		} `xml:"suite"`
		Status struct {
			Text      string `xml:",chardata"`
			Status    string `xml:"status,attr"`
			Starttime string `xml:"starttime,attr"`
			Endtime   string `xml:"endtime,attr"`
		} `xml:"status"`
	} `xml:"suite"`
	Statistics struct {
		Text  string `xml:",chardata"`
		Total struct {
			Text string `xml:",chardata"`
			Stat []struct {
				Text string `xml:",chardata"`
				Pass string `xml:"pass,attr"`
				Fail string `xml:"fail,attr"`
			} `xml:"stat"`
		} `xml:"total"`
		Tag   string `xml:"tag"`
		Suite struct {
			Text string `xml:",chardata"`
			Stat []struct {
				Text string `xml:",chardata"`
				Pass string `xml:"pass,attr"`
				Fail string `xml:"fail,attr"`
				ID   string `xml:"id,attr"`
				Name string `xml:"name,attr"`
			} `xml:"stat"`
		} `xml:"suite"`
	} `xml:"statistics"`
	Errors string `xml:"errors"`
}

type FailedTest struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Suite   string `json:"suite"`
}

const (
	tokenMissingMessage      = "Token missing. Please define GH_ACCESS_TOKEN environment variable."
	ownerMissingMessage      = "Owner missing. Please define REPO_OWNER environment variable."
	shaMissingMessage        = "Commit SHA missing. Please define COMMIT_SHA environment variable."
	repositoryMissingMessage = "Repository missing. Please define REPOSITORY environment variable."
	reportPathMissingMessage = "Report path missing. Please define REPORT_PATH environment variable."
)

var (
	token      = flag.String("access_token", os.Getenv("GH_ACCESS_TOKEN"), "GitHub Access Token")
	owner      = flag.String("repo_owner", os.Getenv("REPO_OWNER"), "Repository owner")
	sha        = flag.String("sha", os.Getenv("COMMIT_SHA"), "Commit`s SHA")
	repository = flag.String("repository", os.Getenv("REPOSITORY"), "Repository")
	reportPath = flag.String("report_path", os.Getenv("REPORT_PATH"), "Location of output.xml")
)

func readOutput() (*os.File, error) {
	path := filepath.Join(*reportPath, "output.xml")
	xmlFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return xmlFile, nil
}

func authenticate() (context.Context, *http.Client) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return ctx, tc
}

func main() {

	// Verify that environment variables are present
	if *token == "" {
		log.Fatal(tokenMissingMessage)
	}

	if *owner == "" {
		log.Fatal(ownerMissingMessage)
	}

	if *sha == "" {
		log.Fatal(shaMissingMessage)
	}

	if *repository == "" {
		log.Fatal(repositoryMissingMessage)
	}

	if *reportPath == "" {
		log.Fatal(reportPathMissingMessage)
	}

	// Read output.xml file to memory
	file, err := readOutput()
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}

	var robot Robot

	var failures []FailedTest

	xml.Unmarshal(byteValue, &robot)

	//for k := 0; k < len(robot.Suite.Suite.Test); k++ {

	//}

	var suite string

	for i := 0; i < len(robot.Suite.Suite.Test); i++ {
		if robot.Suite.Suite.Test[i].Status.Status == "FAIL" {
			for k := 0; k < len(robot.Statistics.Suite.Stat); k++ {
				splittedID := strings.SplitN(robot.Suite.Suite.Test[i].ID, "-", 3)
				suiteID := splittedID[0] + "-" + splittedID[1]
				testSuiteID := robot.Statistics.Suite.Stat[k].ID
				if testSuiteID == suiteID {
					suite = robot.Statistics.Suite.Stat[k].Name
				}
				//fmt.Println(robot.Statistics.Suite.Stat[k].Name)
			}
			name := robot.Suite.Suite.Test[i].Name
			status := strings.ReplaceAll(robot.Suite.Suite.Test[i].Status.Text, "\n", " ")
			input := []byte(fmt.Sprintf(`[{
				"name": "%v",
				"message": "%v",
				"suite": "%v"
			}]`, name, status, suite))
			var tmpFailure []FailedTest
			err := json.Unmarshal(input, &tmpFailure)
			if err != nil {
				fmt.Println(err)
			}

			failures = append(failures, tmpFailure...)
		}
	}

	// Create oauth client
	ctx, tc := authenticate()
	// Use the oauth client to authenticate to Github API
	client := github.NewClient(tc)

	passed := robot.Statistics.Total.Stat[0].Pass
	failed := robot.Statistics.Total.Stat[0].Fail
	total := len(robot.Suite.Suite.Test)

	vars := make(map[string]interface{})
	vars["Passed"] = &passed
	vars["Failed"] = &failed
	vars["Total"] = &total
	vars["FailedTests"] = &failures

	var tp bytes.Buffer

	templatelocation := "/template.txt"
	if err != nil {
		log.Fatal(err)
	}

	tpl, err := template.ParseFiles(templatelocation)

	tpl.Execute(&tp, vars)

	result := tp.String()

	commitComment := &github.RepositoryComment{Body: &result}

	_, _, err = client.Repositories.CreateComment(ctx, *owner, *repository, *sha, commitComment)
	if err != nil {
		log.Fatal(err)
	}

}
