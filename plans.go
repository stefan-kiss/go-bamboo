package bamboo

import (
	"fmt"
	"net/http"
	"strconv"
)

// PlanService handles communication with the plan related methods
type PlanService service

// PlanCreateBranchOptions specifies the optional parameters
// for the CreateBranch method
type PlanCreateBranchOptions struct {
	VCSBranch string
}

// PlanResponse encapsultes a response from the plan service
type PlanResponse struct {
	*ResourceMetadata
	Plans *Plans `json:"plans"`
}

// Plans is a collection of Plan objects
type Plans struct {
	*CollectionMetadata
	PlanList []*Plan `json:"plan"`
}

// Plan is the definition of a single plan
type Plan struct {
	ShortName       string           `json:"shortName,omitempty"`
	ShortKey        string           `json:"shortKey,omitempty"`
	Type            string           `json:"type,omitempty"`
	Enabled         bool             `json:"enabled,omitempty"`
	Link            *Link            `json:"link,omitempty"`
	Key             string           `json:"key,omitempty"`
	Name            string           `json:"name,omitempty"`
	PlanKey         *PlanKey         `json:"planKey,omitempty"`
	VariableContext *VariableContext `json:"variableContext"`
}

// PlanKey holds the plan-key for a plan
type PlanKey struct {
	Key string `json:"key,omitempty"`
}

type VariableContext struct {
	Size       int          `json:"size"`
	MaxResults int          `json:"max-results"`
	StartIndex int          `json:"start-index"`
	Variable   VariableList `json:"variable"`
}

type VariableList []PlanVariable

type PlanVariable struct {
	Key          string `json:"key"`
	Value        string `json:"value"`
	VariableType string `json:"variableType"`
	IsPassword   bool   `json:"isPassword"`
}

// CreateBranch will create a plan branch with the given branch name for the specified build
func (p *PlanService) CreateBranch(planKey, branchName string, options *PlanCreateBranchOptions) (bool, *http.Response, error) {
	var u string
	if !emptyStrings(planKey, branchName) {
		u = fmt.Sprintf("plan/%s/branch/%s.json", planKey, branchName)
	} else {
		return false, nil, &simpleError{"Project key and/or branch name cannot be empty"}
	}

	request, err := p.client.NewRequest(http.MethodPut, u, nil)
	if err != nil {
		return false, nil, err
	}

	if options != nil && options.VCSBranch != "" {
		values := request.URL.Query()
		values.Add("vcsBranch", options.VCSBranch)
		request.URL.RawQuery = values.Encode()
	}

	response, err := p.client.Do(request, nil)
	if err != nil {
		return false, response, err
	}

	if !(response.StatusCode == 200) {
		return false, response, &simpleError{fmt.Sprintf("Create returned %d", response.StatusCode)}
	}

	return true, response, nil
}

// GetNumber returns the number of plans on the Bamboo server
func (p *PlanService) GetNumber() (int, *http.Response, error) {
	request, err := p.client.NewRequest(http.MethodGet, "plan.json", nil)
	if err != nil {
		return 0, nil, err
	}

	// Restrict results to one for speed
	values := request.URL.Query()
	values.Add("max-results", "1")
	request.URL.RawQuery = values.Encode()

	planResp := PlanResponse{}
	response, err := p.client.Do(request, &planResp)
	if err != nil {
		return 0, response, err
	}

	if response.StatusCode != 200 {
		return 0, response, &simpleError{fmt.Sprintf("Getting the number of plans returned %s", response.Status)}
	}

	return planResp.Plans.Size, response, nil
}

// List gets information on all plans
func (p *PlanService) List() ([]*Plan, *http.Response, error) {
	// Get number of plans to set max-results
	numPlans, resp, err := p.GetNumber()
	if err != nil {
		return nil, resp, err
	}

	request, err := p.client.NewRequest(http.MethodGet, "plan.json", nil)
	if err != nil {
		return nil, nil, err
	}

	q := request.URL.Query()
	q.Add("max-results", strconv.Itoa(numPlans))
	request.URL.RawQuery = q.Encode()

	planResp := PlanResponse{}
	response, err := p.client.Do(request, &planResp)
	if err != nil {
		return nil, response, err
	}

	if response.StatusCode != 200 {
		return nil, response, &simpleError{fmt.Sprintf("Getting plan information returned %s", response.Status)}
	}

	return planResp.Plans.PlanList, response, nil
}

// ListKeys get all the plan keys for all build plans on Bamboo
func (p *PlanService) ListKeys() ([]string, *http.Response, error) {
	plans, response, err := p.List()
	if err != nil {
		return nil, response, err
	}
	keys := make([]string, len(plans))

	for i, p := range plans {
		keys[i] = p.Key
	}
	return keys, response, nil
}

// ListNames returns a list of ShortNames of all plans
func (p *PlanService) ListNames() ([]string, *http.Response, error) {
	plans, response, err := p.List()
	if err != nil {
		return nil, response, err
	}
	names := make([]string, len(plans))

	for i, p := range plans {
		names[i] = p.ShortName
	}
	return names, response, nil
}

// NamesMap returns a map[string]string where the PlanKey is the key and the ShortName is the value
func (p *PlanService) NamesMap() (map[string]string, *http.Response, error) {
	plans, response, err := p.List()
	if err != nil {
		return nil, response, err
	}

	planMap := make(map[string]string, len(plans))

	for _, p := range plans {
		planMap[p.Key] = p.ShortName
	}
	return planMap, response, nil
}

// Disable will disable a plan or plan branch
func (p *PlanService) Disable(planKey string) (*http.Response, error) {
	u := fmt.Sprintf("plan/%s/enable", planKey)
	request, err := p.client.NewRequest(http.MethodDelete, u, nil)
	if err != nil {
		return nil, err
	}

	response, err := p.client.Do(request, nil)
	if err != nil {
		return response, err
	}
	return response, nil
}

// GetVars will return a plan's variables
func (p *PlanService) GetVars(planKey string) (VariableList, *http.Response, error) {
	planResp := Plan{}

	u := fmt.Sprintf("plan/%s", planKey)
	request, err := p.client.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	q := request.URL.Query()
	q.Add("expand", "variableContext")
	request.URL.RawQuery = q.Encode()

	response, err := p.client.Do(request, &planResp)
	if err != nil {
		return nil, response, err
	}

	if planResp.VariableContext.MaxResults != planResp.VariableContext.Size {
		return planResp.VariableContext.Variable, response, fmt.Errorf("not all results were returned: %d out of %d",
			planResp.VariableContext.Size,
			planResp.VariableContext.MaxResults)
	}

	return planResp.VariableContext.Variable, response, nil
}

// GetVarValueE returns the variable value or error if it's not found
func (vl VariableList) GetVarValueE(name string) (string, error) {
	for _, v := range vl {
		if v.Key == name {
			return v.Value, nil
		}
	}
	return "", fmt.Errorf("not found")
}

// GetVarValue returns the variable value or empty string if it's not found
func (vl VariableList) GetVarValue(name string) string {
	for _, v := range vl {
		if v.Key == name {
			return ""
		}
	}
	return ""
}

func (vl VariableList) ToMap() map[string]string {
	retMap := make(map[string]string, 0)
	for _, v := range vl {
		retMap[v.Key] = v.Value
	}
	return retMap
}
