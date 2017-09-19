package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"git/inspursoft/board/src/apiserver/service"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/astaxie/beego/logs"

	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

const (
	baseRepoPath    = `/repos`
	jenkinsJobURL   = "http://jenkins:8080/job/{{.JobName}}/buildWithParameters?token={{.Token}}&value={{.Value}}&extras={{.Extras}}&file_name={{.FileName}}"
	jenkinsJobToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
)

var repoServePath = filepath.Join(baseRepoPath, "board_repo_serve")
var repoServeURL = filepath.Join("root@gitserver:", "gitserver", "repos", "board_repo_serve")
var repoPath = filepath.Join(baseRepoPath, "board_repo")

type GitRepoController struct {
	baseController
}

type pushObject struct {
	Items    []string `json:"items"`
	Message  string   `json:"message"`
	JobName  string   `json:"job_name"`
	Value    string   `json:"value"`
	Extras   string   `json:"extras"`
	FileName string   `json:"file_name"`
}

func (g *GitRepoController) Prepare() {
	user := g.getCurrentUser()
	if user == nil {
		g.CustomAbort(http.StatusUnauthorized, "Need to login first.")
		return
	}
	g.currentUser = user
	g.isProjectAdmin = (g.currentUser.ProjectAdmin == 1)
	if !g.isProjectAdmin {
		g.CustomAbort(http.StatusForbidden, "Insufficient privileges for manipulating Git repos.")
		return
	}
}

func (g *GitRepoController) CreateServeRepo() {
	_, err := service.InitBareRepo(repoServePath)
	if err != nil {
		g.CustomAbort(http.StatusInternalServerError, fmt.Sprintf("Failed to initialize serve repo: %+v\n", err))
		return
	}
}

func (g *GitRepoController) InitUserRepo() {
	_, err := service.InitRepo(repoServeURL, repoPath)
	if err != nil {
		g.CustomAbort(http.StatusInternalServerError, fmt.Sprintf("Failed to initialize user's repo: %+v\n", err))
		return
	}

	subPath := g.GetString("sub_path")
	if subPath != "" {
		os.MkdirAll(filepath.Join(repoPath, subPath), 0755)
		if err != nil {
			g.internalError(err)
		}
	}
}

func (g *GitRepoController) PushObjects() {
	var reqPush pushObject
	reqData, err := g.resolveBody()
	if err != nil {
		g.internalError(err)
		return
	}
	err = json.Unmarshal(reqData, &reqPush)
	if err != nil {
		g.internalError(err)
		return
	}

	defaultCommitMessage := fmt.Sprintf("Added items: %s to repo: %s", strings.Join(reqPush.Items, ","), repoPath)

	if len(reqPush.Message) == 0 {
		reqPush.Message = defaultCommitMessage
	}

	repoHandler, err := service.OpenRepo(repoPath)
	if err != nil {
		g.CustomAbort(http.StatusInternalServerError, fmt.Sprintf("Failed to open user's repo: %+v\n", err))
		return
	}
	for _, item := range reqPush.Items {
		repoHandler.Add(item)
	}

	username := g.currentUser.Username
	email := g.currentUser.Email

	_, err = repoHandler.Commit(reqPush.Message, &object.Signature{Name: username, Email: email})
	if err != nil {
		g.CustomAbort(http.StatusInternalServerError, fmt.Sprintf("Failed to commit changes to user's repo: %+v\n", err))
		return
	}
	err = repoHandler.Push()
	if err != nil {
		g.CustomAbort(http.StatusInternalServerError, fmt.Sprintf("Failed to push objects to git repo: %+v\n", err))
	}

	templates := template.Must(template.New("job_url").Parse(jenkinsJobURL))
	var triggerURL bytes.Buffer
	data := struct {
		Token    string
		JobName  string
		Value    string
		Extras   string
		FileName string
	}{
		Token:    jenkinsJobToken,
		JobName:  reqPush.JobName,
		Value:    reqPush.Value,
		Extras:   reqPush.Extras,
		FileName: reqPush.FileName,
	}
	templates.Execute(&triggerURL, data)
	logs.Debug("Jenkins trigger url: %s", triggerURL.String())
	resp, err := http.Get(triggerURL.String())
	if err != nil {
		g.internalError(err)
	}
	g.CustomAbort(resp.StatusCode, "")
}

func (g *GitRepoController) PullObjects() {
	target := g.GetString("target")
	if target == "" {
		g.CustomAbort(http.StatusBadRequest, "No target provided for pulling.")
		return
	}
	targetPath := filepath.Join(baseRepoPath, target)
	repoHandler, err := service.InitRepo(repoServeURL, targetPath)
	if err != nil {
		g.CustomAbort(http.StatusInternalServerError, fmt.Sprintf("Failed to open user's repo: %+v\n", err))
		return
	}
	err = repoHandler.Pull()
	if err != nil {
		g.CustomAbort(http.StatusInternalServerError, fmt.Sprintf("Failed to pull objects from git repo: %+v\n", err))
	}
}

func InternalPushObjects(p *pushObject, g *baseController) (int, string, error) {

	defaultCommitMessage := fmt.Sprintf("Added items: %s to repo: %s", strings.Join(p.Items, ","), repoPath)

	if len(p.Message) == 0 {
		p.Message = defaultCommitMessage
	}

	repoHandler, err := service.OpenRepo(repoPath)
	if err != nil {
		return http.StatusInternalServerError, "Failed to open user's repo", err
	}
	for _, item := range p.Items {
		repoHandler.Add(item)
	}

	username := g.currentUser.Username
	email := g.currentUser.Email

	_, err = repoHandler.Commit(p.Message, &object.Signature{Name: username, Email: email})
	if err != nil {
		return http.StatusInternalServerError, "Failed to commit changes to user's repo", err
	}
	err = repoHandler.Push()
	if err != nil {
		return http.StatusInternalServerError, "Failed to push objects to git repo", err
	}

	templates := template.Must(template.New("job_url").Parse(jenkinsJobURL))
	var triggerURL bytes.Buffer
	data := struct {
		Token    string
		JobName  string
		Value    string
		Extras   string
		FileName string
	}{
		Token:    jenkinsJobToken,
		JobName:  p.JobName,
		Value:    p.Value,
		Extras:   p.Extras,
		FileName: p.FileName,
	}
	templates.Execute(&triggerURL, data)
	logs.Debug("Jenkins trigger url: %s", triggerURL.String())
	resp, err := http.Get(triggerURL.String())
	if err != nil {
		return http.StatusInternalServerError, "Failed to triggerURL", err
	}
	return resp.StatusCode, "Internal Push Object successfully", err
}