package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

var (
	UserConfig *Config

	UnauthorizedError          = errors.New("Unauthorized error returned from server")
	ForbiddenError             = errors.New("Forbidden error returned from server")
	BotNotFoundError           = errors.New("Specified bot is not found")
	JobNotFoundError           = errors.New("Specified job is not found")
	JobAlreadyDoneError        = errors.New("Specified job has already done")
	BotAlreadyRunningError     = errors.New("Specified bot is already running")
	TooManyExecuteRequestError = errors.New("Too many requests error returned from server")
	BotExecutionIsAbortedError = errors.New("Specified bot execution is aborted")
)

func setup() {
	var err error

	UserConfig, err = getConfig()
	if err == ConfigFileNotFoundError {
		UserConfig, err = createConfigFile()
		if err != nil {
			fmt.Fprintf(os.Stderr, "config file create failed.\n%v", err)
			os.Exit(1)
		}
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "config file load error.\n%v", err)
		os.Exit(1)
	}
}

func main() {
	setup()

	var doDisplayProfile bool
	var doReconfigProfile bool
	var doListingBots bool
	var showBotId string
	var listingJobsBotId string
	var abortJobId string
	var formatType string
	var timeoutTime int
	var callbackEndpoint string
	var callbackTries int
	var execInputParam string

	flag.BoolVar(&doDisplayProfile, "p", false, "display profile")
	flag.BoolVar(&doReconfigProfile, "r", false, "display profile")
	flag.BoolVar(&doListingBots, "l", false, "listing bots")
	flag.StringVar(&showBotId, "s", "", "show bot detail")
	flag.StringVar(&listingJobsBotId, "j", "", "listing specify bot jobs")
	flag.StringVar(&abortJobId, "a", "", "abort bot job")
	flag.StringVar(&formatType, "f", "json", "output format type")
	flag.IntVar(&timeoutTime, "t", 0, "timeout time")
	flag.StringVar(&callbackEndpoint, "u", "", "callback endpoint url")
	flag.IntVar(&callbackTries, "T", 0, "number of callback retry trials")
	flag.StringVar(&execInputParam, "i", "", "input parameters for execute bot")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: cbot-cli [OPTION]... [EXECUTE BOT_ID]
  Options:
    -i             : input parameters for execute bot.(ex: key:value,key2:value2...)[default '']
    -t             : timeout time at bot execution.(0-25000)[default 0]
    -u             : callback endpoint url.(needs prefix https://)[default '']
    -T             : number of callback retry trials.(0-5)[default 0]
    -p             : display current config profile.
    -r             : reconfiguration profile.
    -h             : display this help.
    -l             : listing your bots.
    -s BOT_ID      : show specify bot detail.
    -j BOT_ID      : listing specify bot jobs.
    -a JOB_ID      : abort specify bot job.
    -f json | text : output format type.[default 'json'] (support listing options only)
`)
	}

	flag.Parse()
	args := flag.Args()

	if doDisplayProfile == true {
		displayCurrentConfig()
		os.Exit(0)
	}

	if doReconfigProfile == true {
		updateConfigFile()
		os.Exit(0)
	}

	if doListingBots == true {
		listingBotsPortal(formatType)
		os.Exit(0)
	}

	if showBotId != "" {
		showBotPortal(showBotId)
		os.Exit(0)
	}

	if listingJobsBotId != "" {
		listingJobsPortal(listingJobsBotId, formatType)
		os.Exit(0)
	}

	if abortJobId != "" {
		abortJobPortal(abortJobId)
		os.Exit(0)
	}

	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	p := execParameter{
		TimeoutTime:      timeoutTime,
		CallbackEndpoint: callbackEndpoint,
		CallbackTries:    callbackTries,
		execInputParam:   execInputParam,
	}

	execBotPortal(args[0], p)

	os.Exit(0)
}
