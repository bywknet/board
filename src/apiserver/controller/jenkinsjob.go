package controller

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/gorilla/websocket"
)

const jenkinsLastBuildNumberTemplateURL = "http://jenkins:8080/job/{{.JobName}}/lastBuild/buildNumber"
const jenkinsBuildConsoleTemplateURL = "http://jenkins:8080/job/{{.JobName}}/{{.BuildSerialID}}/consoleText"

type jobConsole struct {
	JobName       string `json:"job_name"`
	BuildSerialID string `json:"build_serial_id"`
}

type JenkinsJobController struct {
	baseController
}

func (j *JenkinsJobController) Prepare() {
	user := j.getCurrentUser()
	if user == nil {
		j.customAbort(http.StatusUnauthorized, "Need to login first.")
		return
	}
	j.currentUser = user
	j.isProjectAdmin = (j.currentUser.ProjectAdmin == 1)
	if !j.isProjectAdmin {
		j.customAbort(http.StatusForbidden, "Insufficient privileges for manipulating Git repos.")
		return
	}
}

func generateURL(rawTemplate string, data interface{}) (string, error) {
	t := template.Must(template.New("").Parse(rawTemplate))
	var targetURL bytes.Buffer
	err := t.Execute(&targetURL, data)
	if err != nil {
		return "", err
	}
	return targetURL.String(), nil
}

func (j *JenkinsJobController) Console() {
	jobName := j.GetString("job_name")
	if jobName == "" {
		j.customAbort(http.StatusBadRequest, "No job name found.")
		return
	}

	query := jobConsole{JobName: jobName}

	lastBuildNumberURL, err := generateURL(jenkinsLastBuildNumberTemplateURL, query)
	if err != nil {
		j.internalError(err)
		return
	}

	resp, err := http.Get(lastBuildNumberURL)
	if err != nil {
		j.internalError(err)
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		j.internalError(err)
		return
	}

	lastBuildNumber, err := strconv.Atoi(string(data))
	if err != nil {
		j.internalError(err)
		return
	}

	query.BuildSerialID = j.GetString("build_serial_id", strconv.Itoa(lastBuildNumber+1))

	buildConsoleURL, err := generateURL(jenkinsBuildConsoleTemplateURL, query)
	if err != nil {
		j.internalError(err)
		return
	}

	logs.Debug("Requested Jenkins build console URL: %s", buildConsoleURL)

	ws, err := websocket.Upgrade(j.Ctx.ResponseWriter, j.Ctx.Request, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		j.customAbort(http.StatusBadRequest, "Not a websocket handshake.")
		return
	} else if err != nil {
		j.customAbort(http.StatusInternalServerError, "Cannot setup websocket connection.")
		return
	}
	defer ws.Close()

	req, err := http.NewRequest("GET", buildConsoleURL, nil)
	client := http.Client{}

	buffer := make(chan []byte, 1024)
	done := make(chan bool)

	timer := time.NewTimer(time.Second * 120)
	ticker := time.NewTicker(time.Second * 1)

	go func() {
		for range ticker.C {
			resp, err = client.Do(req)
			if err != nil {
				j.internalError(err)
				return
			}
			data, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				j.internalError(err)
				return
			}
			buffer <- data
			resp.Body.Close()

			for _, line := range strings.Split(string(data), "\n") {
				if strings.HasPrefix(line, "Finished:") {
					ticker.Stop()
					done <- true
				}
			}
		}
	}()

	for {
		select {
		case content := <-buffer:
			err = ws.WriteMessage(websocket.TextMessage, content)
		case <-done:
			err = ws.Close()
			logs.Debug("WS is being closed.")
		case <-timer.C:
			err = ws.Close()
			ticker.Stop()
			logs.Debug("WS is being closed due to timeout.")
		}
		if err != nil {
			logs.Error("Failed to write message: %+v", err)
		}
	}
}
