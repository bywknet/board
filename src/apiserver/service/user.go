package service

import (
	"bytes"
	"errors"
	"fmt"
	"git/inspursoft/board/src/apiserver/service/devops/gogs"
	"git/inspursoft/board/src/common/dao"
	"git/inspursoft/board/src/common/model"
	"git/inspursoft/board/src/common/utils"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/astaxie/beego/logs"
)

var boardHost = utils.GetConfig("BOARD_HOST")
var sshKeyPath = utils.GetConfig("SSH_KEY_PATH")
var gogitsSSHPort = utils.GetConfig("GOGITS_SSH_PORT")

const (
	sshPrivateKey = "id_rsa"
	sshPublicKey  = "id_rsa.pub"
)

func ConfigSSHAccess(username string, accessToken string) error {
	sshKeyUserPath := filepath.Join(sshKeyPath(), username)
	err := os.MkdirAll(sshKeyUserPath, 0755)
	if err != nil {
		return err
	}
	sshPrivateKeyPath := filepath.Join(sshKeyUserPath, sshPrivateKey)
	err = exec.Command("ssh-keygen", "-t", "rsa", "-b", "4096", "-f", sshPrivateKeyPath, "-q", "-N", "").Run()
	if err != nil {
		return fmt.Errorf("Failed to generate SSH Key pairs: %+v", err)
	}
	err = exec.Command("ssh", "-i", sshPrivateKeyPath, "-4", boardHost(), "-o", "StrictHostKeyChecking=no", "-p", gogitsSSHPort()).Run()
	if err != nil {
		logs.Warn("Failed to add Public key to known hosts: %+v", err)
	}
	data, err := ioutil.ReadFile(filepath.Join(sshKeyUserPath, sshPublicKey))
	if err != nil {
		return err
	}
	publicKey := bytes.NewBuffer(data).String()
	return gogs.NewGogsHandler(username, accessToken).CreatePublicKey(fmt.Sprintf("%s's access public key", username), publicKey)
}

func SignUp(user model.User) (bool, error) {
	err := gogs.SignUp(user)
	if err != nil {
		return false, fmt.Errorf("Failed to create Gogs account for DevOps: %+v", err)
	}
	accessToken, err := gogs.CreateAccessToken(user.Username, user.Password)
	if err != nil {
		return false, err
	}
	err = ConfigSSHAccess(user.Username, accessToken.Sha1)
	if err != nil {
		return false, err
	}
	user.RepoToken = accessToken.Sha1
	user.Salt = utils.GenerateRandomString()
	user.Password = utils.Encrypt(user.Password, user.Salt)
	userID, err := dao.AddUser(user)
	if err != nil {
		return false, err
	}
	return (userID != 0), nil
}

func GetUserByID(userID int64) (*model.User, error) {
	query := model.User{ID: userID, Deleted: 0}
	user, err := dao.GetUser(query, "id", "deleted")
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUsers(field string, value interface{}, selectedFields ...string) ([]*model.User, error) {
	return dao.GetUsers(field, value, selectedFields...)
}

func GetPaginatedUsers(field string, value interface{}, pageIndex int, pageSize int, selectedField ...string) (*model.PaginatedUsers, error) {
	return dao.GetPaginatedUsers(field, value, pageIndex, pageSize, selectedField...)
}

func UpdateUser(user model.User, selectedFields ...string) (bool, error) {
	if user.ID == 0 {
		return false, errors.New("No user ID provided.")
	}
	_, err := dao.UpdateUser(user, selectedFields...)
	if err != nil {
		return false, err
	}
	return true, nil
}

func DeleteUser(userID int64) (bool, error) {
	user := model.User{ID: userID, Deleted: 1}
	_, err := dao.UpdateUser(user, "deleted")
	if err != nil {
		return false, err
	}
	return true, nil
}

func UserExists(fieldName string, value string, userID int64) (bool, error) {
	query := model.User{ID: userID, Username: value, Email: value}
	user, err := dao.GetUser(query, fieldName)
	if err != nil {
		return false, err
	}
	if userID == 0 {
		return (user != nil && user.ID != 0), nil
	}
	return (user != nil && user.ID != userID), nil
}

func IsSysAdmin(userID int64) (bool, error) {
	query := model.User{ID: userID}
	user, err := dao.GetUser(query, "id")
	if err != nil {
		return false, err
	}
	return (user != nil && user.ID != 0 && user.SystemAdmin == 1), nil
}
