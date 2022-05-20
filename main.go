package main

import (
	"bytes"
	"context"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type Test struct {
	Name          string
	Status        string
	Suite         string
	ExecutionTime float64
	Message       string
}

type Statistics struct {
	Pass string
	Fail string
	Skip string
}

const (
	tokenMissingMessage      = "Token missing. Please define GH_ACCESS_TOKEN environment variable."
	ownerMissingMessage      = "Owner missing. Please define REPO_OWNER environment variable."
	shaMissingMessage        = "Either SHA or PR ID needs to be defined. Please define COMMIT_SHA or PR_ID environment variable."
	repositoryMissingMessage = "Repository missing. Please define REPOSITORY environment variable."
	reportPathMissingMessage = "Report path missing. Please define REPORT_PATH environment variable."
)

var (
	token           = flag.String("access_token", os.Getenv("GH_ACCESS_TOKEN"), "GitHub Access Token")
	owner           = flag.String("repo_owner", os.Getenv("REPO_OWNER"), "Repository owner")
	sha             = flag.String("sha", os.Getenv("COMMIT_SHA"), "Commit`s SHA")
	repository      = flag.String("repository", os.Getenv("REPOSITORY"), "Repository")
	reportPath      = flag.String("report_path", os.Getenv("REPORT_PATH"), "Location of output.xml")
	pullRequestID   = flag.String("pull_request_id", os.Getenv("PR_ID"), "ID of pull")
	summary         = flag.String("summary", os.Getenv("SUMMARY"), "If true show report in job summary")
	onlySummary     = flag.String("only_summary", os.Getenv("ONLY_SUMMARY"), "If true only add report to job summary")
	showPassedTests = flag.String("show_passed_tests", os.Getenv("SHOW_PASSED_TESTS"), "If true display also passed tests")
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

	if *onlySummary == "true" && *summary != "true" {
		*summary = "true"
	}

	if *onlySummary != "true" {
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

	output, err := xmlquery.Parse(file)
	if err != nil {
		log.Fatal(err)
	}

	tests := xmlquery.Find(output, "//test")

	var failedTests []Test
	var passedTests []Test

	for _, test := range tests {
		name := test.SelectAttr("name")
		status := test.SelectElement("status").SelectAttr("status")
		suite := test.Parent.SelectAttr("name")
		startTime := test.SelectElement("status").SelectAttr("starttime")
		endTime := test.SelectElement("status").SelectAttr("endtime")
		message := strings.ReplaceAll(test.SelectElement("status").InnerText(), "\n", " ")

		executionTime, err := getExecutionTime(startTime, endTime)
		if err != nil {
			log.Println(err)
		}

		if status == "PASS" {
			passedTests = append(passedTests, Test{
				Name:          name,
				Status:        status,
				Suite:         suite,
				ExecutionTime: executionTime,
				Message:       message,
			})
		}

		if status == "FAIL" {
			failedTests = append(failedTests, Test{
				Name:          name,
				Status:        status,
				Suite:         suite,
				ExecutionTime: executionTime,
				Message:       message,
			})
		}
	}

	statistics := xmlquery.FindOne(output, "//statistics//total//stat")

	statistic := Statistics{
		Pass: statistics.SelectAttr("pass"),
		Fail: statistics.SelectAttr("fail"),
		Skip: statistics.SelectAttr("skip"),
	}

	passInt, err := strconv.Atoi(statistic.Pass)
	if err != nil {
		log.Fatal(err)
	}

	failInt, err := strconv.Atoi(statistic.Fail)
	if err != nil {
		log.Fatal(err)
	}

	vars := make(map[string]interface{})
	vars["Passed"] = statistic.Pass
	vars["Failed"] = statistic.Fail
	vars["Skipped"] = statistic.Skip
	vars["Total"] = passInt + failInt
	vars["PassPercentage"] = passPercentage(passInt, failInt)
	vars["PassedTests"] = passedTests
	vars["FailedTests"] = failedTests
	vars["ShowPassedTests"] = *showPassedTests

	templatelocation := "./assets/template.txt"
	tpl, err := template.ParseFiles(templatelocation)
	if err != nil {
		log.Fatal(err)
	}

	var tp bytes.Buffer

	tpl.Execute(&tp, vars)

	result := tp.String()

	// Create oauth client
	ctx, tc := authenticate()
	// Use the oauth client to authenticate to Github API
	client := github.NewClient(tc)

	if *onlySummary != "true" {
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
