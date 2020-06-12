package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

type execParameter struct {
	TimeoutTime      int               `json:"timeout_time"`
	CallbackEndpoint string            `json:"callback_endpoint"`
	CallbackTries    int               `json:"callback_tries"`
	Input            map[string]string `json:"input"`
	execInputParam   string
}

type execBotResponse struct {
	Code    int    `json:"code"`
	JobId   string `json:"job_id"`
	BotId   string `json:"bot_id"`
	BotName string `json:"bot_name"`
	Status  int    `json:"status"`
}

func setupParameter(param *execParameter) error {
	param.Input = make(map[string]string)
	pairs := strings.Split(param.execInputParam, ",")

	for _, pair := range pairs {
		v := strings.Split(pair, ":")
		if len(v)%2 != 0 {
			return errors.New("invalidate format parameter.")
		}
		param.Input[v[0]] = v[1]
	}
	return nil
}

func execBotPortal(botId string, param execParameter) {
	err := setupParameter(&param)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parameter format is invalidate. Ex: key:value")
		os.Exit(1)
	}
	err = execBot(botId, param)
	if err == UnauthorizedError {
		fmt.Fprintf(os.Stderr, "unauthorized error returned. Check your access token and key.")
		os.Exit(1)
	} else if err == ForbiddenError {
		fmt.Fprintf(os.Stderr, "forbidden error returned. Do you have a bot execute authorize?")
		os.Exit(1)
	} else if err == BotNotFoundError {
		fmt.Fprintf(os.Stderr, "bot id '%s' is not found.", botId)
		os.Exit(1)
		/*
			Hmm...Cloud Bot always return 202 status?
				} else if err == BotAlreadyRunningError {
					fmt.Fprintf(os.Stderr, "bot id '%s' is already running.", botId)
					os.Exit(1)
		*/
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

func buildExecBotURL(botId string) (string, error) {
	u, err := url.Parse(UserConfig.ApiPath)
	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, "bots", botId, "jobs")

	return u.String(), nil
}

func buildExecBotRequest(botId string, param execParameter) (*http.Request, error) {
	url, err := buildExecBotURL(botId)
	if err != nil {
		return nil, err
	}

	p, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(p))
	if err != nil {
		return nil, err
	}

	req.Header.Add("content-type", "application/json")
	req.Header.Add("content-language", UserConfig.ContentLanguage)
	req.Header.Add("access-token", UserConfig.AccessToken)
	req.Header.Add("secret-key", UserConfig.SecretKey)

	return req, nil
}

func processExecBotResponse(resp *http.Response) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var ret execBotResponse
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return err
	}

	switch ret.Code {
	case 200:
		// nothing todo
	case 202:
		// nothing todo
		//return BotAlreadyRunningError
	case 401:
		return UnauthorizedError
	case 403:
		return ForbiddenError
	case 404:
		return BotNotFoundError
	case 410:
		return BotExecutionIsAbortedError
	case 429:
		return TooManyExecuteRequestError
	default:
		// error
		return fmt.Errorf("response code '%d' returned.", ret.Code)
	}

	fmt.Println(string(body))

	return nil
}

func execBot(botId string, param execParameter) error {
	client := http.DefaultClient

	req, err := buildExecBotRequest(botId, param)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := processExecBotResponse(resp); err != nil {
		return err
	}

	return nil
}
