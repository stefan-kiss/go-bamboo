package bamboo

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	PlanExpandResults           = "stages.stage.results.result"
	PlanExpandVariables         = "variables"
	PlanExpandArtifacts         = "artifacts"
	PlanExpandFailedTestResults = "stages.stage.results.result.testResults.failedTests.testResult.errors"
)

// ResultService handles communication with build results
type ResultService service

// ResultsResponse encapsulates the information from
// requesting result information
type ResultsResponse struct {
	*ResourceMetadata
	Results *Results `json:"results"`
}

// Results is the collection of results
type Results struct {
	CollectionMetadata `json:",inline"`
	Result             []Result `json:"result"`
}

type Stage struct {
	Expand             string  `json:"expand"`
	Name               string  `json:"name"`
	Description        string  `json:"description"`
	Id                 int     `json:"id"`
	LifeCycleState     string  `json:"lifeCycleState"`
	State              string  `json:"state"`
	DisplayClass       string  `json:"displayClass"`
	DisplayMessage     string  `json:"displayMessage"`
	CollapsedByDefault bool    `json:"collapsedByDefault"`
	Manual             bool    `json:"manual"`
	Restartable        bool    `json:"restartable"`
	Runnable           bool    `json:"runnable"`
	Results            Results `json:"results"`
}

type Stages struct {
	CollectionMetadata `json:",inline"`
	Stage              []Stage `json:"stage"`
}

type Artifact struct {
	Name                  string `json:"name"`
	Link                  Link   `json:"link"`
	ProducerJobKey        string `json:"producerJobKey"`
	Shared                bool   `json:"shared"`
	Size                  int    `json:"size"`
	PrettySizeDescription string `json:"prettySizeDescription"`
}

type Artifacts struct {
	CollectionMetadata `json:",inline"`
	Artifact           []Artifact `json:"artifact"`
}

type ErrorMessage struct {
	Message string `json:"message"`
}

type TestError struct {
	CollectionMetadata `json:",inline"`
	Error              []ErrorMessage `json:"error"`
}
type TestDetail struct {
	TestCaseId        int       `json:"testCaseId"`
	ClassName         string    `json:"className"`
	MethodName        string    `json:"methodName"`
	Status            string    `json:"status"`
	Duration          int       `json:"duration"`
	DurationInSeconds int       `json:"durationInSeconds"`
	Errors            TestError `json:"errors"`
}
type TestResult struct {
	CollectionMetadata `json:",inline"`
	TestResult         []TestDetail `json:"testResult"`
}

type TestResults struct {
	Expand              string     `json:"expand"`
	All                 int        `json:"all"`
	Successful          int        `json:"successful"`
	Failed              int        `json:"failed"`
	NewFailed           int        `json:"newFailed"`
	ExistingFailed      int        `json:"existingFailed"`
	Fixed               int        `json:"fixed"`
	Quarantined         int        `json:"quarantined"`
	Skipped             int        `json:"skipped"`
	AllTests            TestResult `json:"allTests"`
	SuccessfulTests     TestResult `json:"successfulTests"`
	FailedTests         TestResult `json:"failedTests"`
	NewFailedTests      TestResult `json:"newFailedTests"`
	ExistingFailedTests TestResult `json:"existingFailedTests"`
	FixedTests          TestResult `json:"fixedTests"`
	QuarantinedTests    TestResult `json:"quarantinedTests"`
	SkippedTests        TestResult `json:"skippedTests"`
}

// Result represents all the information associated with a build result
type Result struct {
	ChangeSet              `json:"changes"`
	ID                     int         `json:"id"`
	PlanName               string      `json:"planName"`
	ProjectName            string      `json:"projectName"`
	BuildResultKey         string      `json:"buildResultKey"`
	LifeCycleState         string      `json:"lifeCycleState"`
	BuildStartedTime       string      `json:"buildStartedTime"`
	BuildCompletedTime     string      `json:"buildCompletedTime"`
	BuildDurationInSeconds int         `json:"buildDurationInSeconds"`
	VcsRevisionKey         string      `json:"vcsRevisionKey"`
	BuildTestSummary       string      `json:"buildTestSummary"`
	SuccessfulTestCount    int         `json:"successfulTestCount"`
	FailedTestCount        int         `json:"failedTestCount"`
	QuarantinedTestCount   int         `json:"quarantinedTestCount"`
	SkippedTestCount       int         `json:"skippedTestCount"`
	Finished               bool        `json:"finished"`
	Successful             bool        `json:"successful"`
	BuildReason            string      `json:"buildReason"`
	ReasonSummary          string      `json:"reasonSummary"`
	Key                    string      `json:"key"`
	State                  string      `json:"state"`
	BuildState             string      `json:"buildState"`
	Number                 int         `json:"number"`
	BuildNumber            int         `json:"buildNumber"`
	Stages                 Stages      `json:"stages"`
	LogFiles               []string    `json:"logFiles"`
	Artifacts              Artifacts   `json:"artifacts"`
	Master                 Plan        `json:"master"`
	Plan                   Plan        `json:"plan"`
	TestResults            TestResults `json:"testResults"`
}

// ChangeSet represents a collection of type Change
type ChangeSet struct {
	Set []Change `json:"change"`
}

// Change represents the author and commit hash of a source code change
type Change struct {
	Author      string `json:"author"`
	ChangeSetID string `json:"changesetId"`
}

// LatestResult returns the latest result information for the given plan key
func (r *ResultService) LatestResult(key string) (*Result, *http.Response, error) {
	result, resp, err := r.NumberedResult(key + "-latest")
	return result, resp, err
}

// NumberedResult returns the result information for the given plan key which includes the build number of the desired result
func (r *ResultService) NumberedResult(key string) (*Result, *http.Response, error) {
	request, err := r.client.NewRequest(http.MethodGet, numberedResultURL(key), nil)
	if err != nil {
		return nil, nil, err
	}

	result := Result{}
	response, err := r.client.Do(request, &result)
	if err != nil {
		return nil, response, err
	}

	if response.StatusCode != 200 {
		return nil, response, &simpleError{fmt.Sprintf("API returned unexpected status code %d", response.StatusCode)}
	}

	return &result, response, err
}

// ListResults lists the results for a plan
func (r *ResultService) ListResults(key string) ([]Result, *http.Response, error) {
	request, err := r.client.NewRequest(http.MethodGet, listResultsURL(key), nil)
	if err != nil {
		return nil, nil, err
	}

	result := ResultsResponse{}
	response, err := r.client.Do(request, &result)
	if err != nil {
		return nil, response, err
	}

	if response.StatusCode != 200 {
		return nil, response, &simpleError{fmt.Sprintf("API returned unexpected status code %d", response.StatusCode)}
	}

	return result.Results.Result, response, err
}

func (r *ResultService) GetExpanded(key string, expand []string) (*Result, *http.Response, error) {

	pathStr := fmt.Sprintf("result/%s", key)
	request, err := r.client.NewRequest(http.MethodGet, pathStr, nil)
	if err != nil {
		return nil, nil, err
	}

	expandQ := strings.Join(expand, ",")
	q := request.URL.Query()
	q.Set("expand", expandQ)
	q.Set("includeAllStates", "true")
	request.URL.RawQuery = q.Encode()

	result := Result{}
	response, err := r.client.Do(request, &result)
	if err != nil {
		return nil, response, err
	}

	if response.StatusCode != 200 {
		return nil, response, &simpleError{fmt.Sprintf("API returned unexpected status code %d", response.StatusCode)}
	}

	return &result, response, err

}

func (r *ResultService) GetLatestExpanded(key string, expand []string) (*Result, *http.Response, error) {

	pathStr := fmt.Sprintf("result/%s-latest", key)
	request, err := r.client.NewRequest(http.MethodGet, pathStr, nil)
	if err != nil {
		return nil, nil, err
	}

	expandQ := strings.Join(expand, ",")
	q := request.URL.Query()
	q.Set("expand", expandQ)
	q.Set("includeAllStates", "true")
	request.URL.RawQuery = q.Encode()

	result := Result{}
	response, err := r.client.Do(request, &result)
	if err != nil {
		return nil, response, err
	}

	if response.StatusCode != 200 {
		return nil, response, &simpleError{fmt.Sprintf("API returned unexpected status code %d", response.StatusCode)}
	}

	return &result, response, err

}
