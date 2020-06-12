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

type abortJobResponse struct {
	Code        int    `json:"code"`
	JobId       string `json:"job_id"`
	BotId       string `json:"bot_id"`
	BotName     string `json:"bot_name"`
	Status      int    `json:"status"`
	StartTime   string `json:"start_time"`
	ElapsedTime int    `json:"elapsed_time"`
	Callback    bool   `json:"callback"`
	Message     string `json:"message"`
}

func abortJobPortal(jobId string) {
	err := execAbortJob(jobId)
	if err == UnauthorizedError {
		fmt.Fprintf(os.Stderr, "unauthorized error returned. Check your access token and key.")
		os.Exit(1)
	} else if err == ForbiddenError {
		fmt.Fprintf(os.Stderr, "forbidden error returned. Do you have a job abort authorize?")
		os.Exit(1)
	} else if err == JobNotFoundError {
		fmt.Fprintf(os.Stderr, "bot id '%s' is not found.", jobId)
		os.Exit(1)
	} else if err == JobAlreadyDoneError {
		fmt.Fprintf(os.Stderr, "job id '%s' has already done.", jobId)
		os.Exit(1)
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

func buildAbortJobURL(jobId string) (string, error) {
	u, err := url.Parse(UserConfig.ApiPath)
	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, "jobs", jobId)

	return u.String(), nil
}

func buildAbortJobRequest(botId string) (*http.Request, error) {
	url, err := buildAbortJobURL(botId)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("content-type", "application/json")
	req.Header.Add("content-language", UserConfig.ContentLanguage)
	req.Header.Add("access-token", UserConfig.AccessToken)
	req.Header.Add("secret-key", UserConfig.SecretKey)

	return req, nil
}

func processAbortJobResponse(resp *http.Response) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var ret abortJobResponse
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
	case 404:
		// not found
		return JobNotFoundError
	case 410:
		// already done
		return JobAlreadyDoneError
	default:
		// error
		return fmt.Errorf("response code '%d' returned.", ret.Code)
	}

	return nil
}

func execAbortJob(jobId string) error {
	client := http.DefaultClient

	req, err := buildAbortJobRequest(jobId)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := processAbortJobResponse(resp); err != nil {
		return err
	}

	return nil
}
