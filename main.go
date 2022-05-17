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
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const (
	tokenMissingMessage      = "Token missing. Please define GH_ACCESS_TOKEN environment variable."
	ownerMissingMessage      = "Owner missing. Please define REPO_OWNER environment variable."
	shaMissingMessage        = "Either SHA or PR ID needs to be defined. Please define COMMIT_SHA or PR_ID environment variable."
	repositoryMissingMessage = "Repository missing. Please define REPOSITORY environment variable."
	reportPathMissingMessage = "Report path missing. Please define REPORT_PATH environment variable."
)

var (
	token         = flag.String("access_token", os.Getenv("GH_ACCESS_TOKEN"), "GitHub Access Token")
	owner         = flag.String("repo_owner", os.Getenv("REPO_OWNER"), "Repository owner")
	sha           = flag.String("sha", os.Getenv("COMMIT_SHA"), "Commit`s SHA")
	repository    = flag.String("repository", os.Getenv("REPOSITORY"), "Repository")
	reportPath    = flag.String("report_path", os.Getenv("REPORT_PATH"), "Location of output.xml")
	pullRequestID = flag.String("pull_request_id", os.Getenv("PR_ID"), "ID of pull")
	summary       = flag.String("summary", os.Getenv("SUMMARY"), "If true show report in job summary")
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

	if *sha == "" && *pullRequestID == "" {
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
		log.Fatal(err)
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	var output Output

	xml.Unmarshal(byteValue, &output)

	var passedTests []Test

	var failedTests []Test

	for i := 0; i < len(output.Suite.Suite.Test); i++ {
		if output.Suite.Suite.Test[i].Status.Status == "FAIL" {
			suite := getSuiteName(i, output)
			name := output.Suite.Suite.Test[i].Name
			status := strings.ReplaceAll(output.Suite.Suite.Test[i].Status.Text, "\n", " ")
			testExecutionStartTime := output.Suite.Suite.Test[i].Status.Starttime
			testExecutionEndTime := output.Suite.Suite.Test[i].Status.Endtime

			executionTime, err := getExecutionTime(testExecutionStartTime, testExecutionEndTime)
			if err != nil {
				log.Fatal(err)
			}

			input := []byte(fmt.Sprintf(`[{
				"name": "%v",
				"message": "%v",
				"executionTime": "%v",
				"suite": "%v"
			}]`, name, status, executionTime, suite))
			var tmpFailure []Test
			err = json.Unmarshal(input, &tmpFailure)
			if err != nil {
				log.Fatal(err)
			}

			failedTests = append(failedTests, tmpFailure...)
		}

		if output.Suite.Suite.Test[i].Status.Status == "PASS" {
			suite := getSuiteName(i, output)
			name := output.Suite.Suite.Test[i].Name
			status := strings.ReplaceAll(output.Suite.Suite.Test[i].Status.Text, "\n", " ")
			testExecutionStartTime := output.Suite.Suite.Test[i].Status.Starttime
			testExecutionEndTime := output.Suite.Suite.Test[i].Status.Endtime

			executionTime, err := getExecutionTime(testExecutionStartTime, testExecutionEndTime)
			if err != nil {
				log.Fatal(err)
			}

			input := []byte(fmt.Sprintf(`[{
				"name": "%v",
				"message": "%v",
				"executionTime": "%v",
				"suite": "%v"
			}]`, name, status, executionTime, suite))
			var tmpPass []Test
			err = json.Unmarshal(input, &tmpPass)
			if err != nil {
				log.Fatal(err)
			}

			passedTests = append(passedTests, tmpPass...)
		}
	}

	// Create oauth client
	ctx, tc := authenticate()
	// Use the oauth client to authenticate to Github API
	client := github.NewClient(tc)

	passedInt, err := strconv.Atoi(output.Statistics.Total.Stat[0].Pass)
	if err != nil {
		log.Panic(err)
	}

	passPercentage := fmt.Sprintf("%.2f", float64(passedInt)/float64(len(output.Suite.Suite.Test))*100)

	vars := make(map[string]interface{})
	vars["Passed"] = output.Statistics.Total.Stat[0].Pass
	vars["Failed"] = output.Statistics.Total.Stat[0].Fail
	vars["Skipped"] = output.Statistics.Total.Stat[0].Skip
	vars["Total"] = len(output.Suite.Suite.Test)
	vars["PassPercentage"] = &passPercentage
	vars["PassedTests"] = &passedTests
	vars["FailedTests"] = &failedTests

	var tp bytes.Buffer

	templatelocation := "./assets/template.txt"
	if err != nil {
		log.Fatal(err)
	}

	tpl, err := template.ParseFiles(templatelocation)
	if err != nil {
		log.Fatal(err)
	}

	tpl.Execute(&tp, vars)

	result := tp.String()

	// GitHub's REST API v3 considers every pull request an issue
	if *pullRequestID != "" {
		pullRequestComment := &github.IssueComment{Body: &result}
		updated := false
		convPullRequestID, err := strconv.Atoi(*pullRequestID)
		if err != nil {
			log.Fatal(err)
		}

		// Get all comments in PR
		comments, _, err := client.Issues.ListComments(ctx, *owner, *repository, convPullRequestID, &github.IssueListCommentsOptions{})
		if err != nil {
			log.Fatal(err)
		}

		// Loop through comments and check if report comment already exists
		for _, val := range comments {
			if resultCommentExists(*val.Body) {
				_, _, err = client.Issues.EditComment(ctx, *owner, *repository, *val.ID, pullRequestComment)
				if err != nil {
					log.Fatal(err)
				}
				updated = true
			}
		}

		if !updated {
			_, _, err = client.Issues.CreateComment(ctx, *owner, *repository, convPullRequestID, pullRequestComment)
			if err != nil {
				log.Fatal(err)
			}
		}

	}

	if *pullRequestID == "" {
		commitComment := &github.RepositoryComment{Body: &result}
		_, _, err = client.Repositories.CreateComment(ctx, *owner, *repository, *sha, commitComment)
		if err != nil {
			log.Fatal(err)
		}
	}

	if *summary == "true" {
		github_env := os.Getenv("GITHUB_STEP_SUMMARY")

		file, err = os.OpenFile(github_env, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Println(err)
		}
		defer file.Close()
		if _, err := file.WriteString(result); err != nil {
			log.Println(err)
		}
	}
}
