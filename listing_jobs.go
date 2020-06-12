package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
)

const (
	MAX_LISTING_JOBS = "1000"

	JobStatusExit    = 0
	JobStatusError   = 1
	JobStatusRunning = 2
)

type listingJobsResponse struct {
	Code int                      `json:"code"`
	Jobs []listingJobsResponseJob `json:"jobs"`
}

type listingJobsResponseJob struct {
	JobId       string `json:"job_id"`
	BotId       string `json:"bot_id"`
	BotName     string `json:"bot_name"`
	Status      int    `json:"status"`
	StartTime   string `json:"start_time"`
	ElapsedTime int    `json:"elapsed_time"`
}

func (rj *listingJobsResponseJob) StatusString() string {
	switch rj.Status {
	case JobStatusExit:
		return "exit"
	case JobStatusError:
		return "error"
	case JobStatusRunning:
		return "running"
	default:
		return "???"
	}
}

func listingJobsPortal(botId string, format string) {
	err := execListingJobs(botId, format)
	if err == UnauthorizedError {
		fmt.Fprintf(os.Stderr, "unauthorized error returned. Check your access token and key.")
		os.Exit(1)
	} else if err == ForbiddenError {
		fmt.Fprintf(os.Stderr, "forbidden error returned. Do you have a reference authorize?")
		os.Exit(1)
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

func buildListingJobsURL(botId string) (string, error) {
	u, err := url.Parse(UserConfig.ApiPath)
	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, "bots", botId, "jobs")

	q := u.Query()
	q.Set("limit", MAX_LISTING_JOBS)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func buildListingJobsRequest(botId string) (*http.Request, error) {
	url, err := buildListingJobsURL(botId)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("content-type", "application/json")
	req.Header.Add("content-language", UserConfig.ContentLanguage)
	req.Header.Add("access-token", UserConfig.AccessToken)
	req.Header.Add("secret-key", UserConfig.SecretKey)

	return req, nil
}

func processListingJobsResponse(resp *http.Response, format string) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var ret listingJobsResponse
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return err
	}

	switch ret.Code {
	case 200:
		// nothing todo
	case 401:
		// unauthorized
		return UnauthorizedError
	case 403:
		// forbidden
		return ForbiddenError
	default:
		// error
		return fmt.Errorf("response code '%d' returned.", ret.Code)
	}

	if format != "text" {
		fmt.Println(string(body))
		return nil
	}

	fmt.Println("job_id\tbot_id\tbot_name\tstatus\tstart_time\telapsed_time")
	for _, r := range ret.Jobs {
		fmt.Printf("%s\t%s\t%s\t%s\t%s\t%d\n", r.JobId, r.BotId, r.BotName, r.StatusString(), r.StartTime, r.ElapsedTime)
	}

	return nil
}

func execListingJobs(botId string, format string) error {
	client := http.DefaultClient

	req, err := buildListingJobsRequest(botId)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := processListingJobsResponse(resp, format); err != nil {
		return err
	}

	return nil
}
