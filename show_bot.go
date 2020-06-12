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

type showBotResponse struct {
	Code         int    `json:"code"`
	Id           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Created      string `json:"created"`
	LastModified string `json:"last_modified"`
	Creator      string `json:"creator"`
}

func showBotPortal(botId string) {
	err := execShowBot(botId)
	if err == UnauthorizedError {
		fmt.Fprintf(os.Stderr, "unauthorized error returned. Check your access token and key.")
		os.Exit(1)
	} else if err == ForbiddenError {
		fmt.Fprintf(os.Stderr, "forbidden error returned. Do you have a reference authorize?")
		os.Exit(1)
	} else if err == BotNotFoundError {
		fmt.Fprintf(os.Stderr, "bot id '%s' is not found.", botId)
		os.Exit(1)
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

func buildShowBotURL(botId string) (string, error) {
	u, err := url.Parse(UserConfig.ApiPath)
	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, "bots", botId)

	q := u.Query()
	q.Set("properties", "created,last_modified,creator,input,output")
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func buildShowBotRequest(botId string) (*http.Request, error) {
	url, err := buildShowBotURL(botId)
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

func processShowBotResponse(resp *http.Response) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var ret showBotResponse
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
		return BotNotFoundError
	default:
		// error
		return fmt.Errorf("response code '%d' returned.", ret.Code)
	}

	fmt.Println(string(body))

	return nil
}

func execShowBot(botId string) error {
	client := http.DefaultClient

	req, err := buildShowBotRequest(botId)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := processShowBotResponse(resp); err != nil {
		return err
	}

	return nil
}
