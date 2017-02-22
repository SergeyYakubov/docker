package main

import (
	"encoding/json"
	"regexp"

	"./utils"
	dockerapi "github.com/docker/docker/api"
	dockerclient "github.com/docker/engine-api/client"
	"github.com/docker/go-plugins-helpers/authorization"
	"github.com/pkg/errors"
)

func newPlugin(dockerHost string) (*novolume, error) {
	client, err := dockerclient.NewClient(dockerHost, dockerapi.DefaultVersion, nil, nil)
	if err != nil {
		return nil, err
	}
	return &novolume{client: client}, nil
}

type novolume struct {
	client *dockerclient.Client
}

func isRequestFromRoot(req authorization.Request) bool {
	return req.User == "0:0"
}

func checkBinds(options map[string]interface{}) error {

	b, ok := options["Binds"]
	if !ok {
		return errors.New("Bad Binds flag")
	}

	if b == nil {
		return nil
	}

	binds, ok := b.([]interface{})
	if !ok {
		return errors.New("Bad Binds flag")
	}

	for _, v := range binds {
		str, ok := v.(string)
		if !ok {
			return errors.New("Bad Binds flag")
		}

		if err := utils.CheckBind(str); err != nil {
			return err
		}

	}

	return nil
}

func processStartContainer(req authorization.Request) authorization.Response {

	var docker_options map[string]interface{}

	err := json.Unmarshal(req.RequestBody, &docker_options)

	if err != nil {
		return authorization.Response{Err: "Bad request body"}
	}

	if isRequestFromRoot(req) {
		return authorization.Response{Allow: true}
	}

	host_options, ok := docker_options["HostConfig"].(map[string]interface{})
	if !ok {
		return authorization.Response{Err: "Bad host options"}
	}

	if err := checkBinds(host_options); err != nil {
		return authorization.Response{Err: err.Error()}
	}

	mode, ok := host_options["UsernsMode"]
	if !ok {
		return authorization.Response{Err: "Bad UsernsMode flag"}
	}

	if mode == "" {
		return authorization.Response{Allow: true}
	}

	user, ok := docker_options["User"].(string)
	if !ok || user == "" {
		return authorization.Response{Err: "User not specified. Use -u flag."}
	}

	if !utils.UsersEqual(user, req.User) {
		return authorization.Response{Err: "Not allowed for user " + user + ". Did you set group id?"}
	}

	groups, ok := host_options["GroupAdd"].([]interface{})
	if ok {
		for _, v := range groups {
			switch g := v.(type) {
			case string:
				if !utils.UserGroup(user, g) {
					return authorization.Response{Err: "Wrong user group " + g}
				}
			}
		}
	}

	secopts, ok := host_options["SecurityOpt"]
	if !ok {
		return authorization.Response{Err: "Bad SecurityOpt flag"}
	}

	list, ok := secopts.([]interface{})
	if !ok {
		return authorization.Response{Err: "Bad SecurityOpt flag"}
	}

	for _, v := range list {
		switch str := v.(type) {
		case string:
			if str == "no-new-privileges" {
				return authorization.Response{Allow: true}
			}
		}

	}

	return authorization.Response{Err: "Use --security-opt no-new-privileges falg "}

}

func processContainerCommand(req authorization.Request, cName string) authorization.Response {
	if utils.UserCreatedContainer(req.User, cName) {
		return authorization.Response{Allow: true}
	} else {
		return authorization.Response{Msg: "Command not allowed for this user or resource does not exist"}
	}
}

func processExecContainer(req authorization.Request, cName string) authorization.Response {

	var docker_options map[string]interface{}
	err := json.Unmarshal(req.RequestBody, &docker_options)

	if err != nil {
		return authorization.Response{Err: "Bad Request Body"}
	}

	if isRequestFromRoot(req) {
		return authorization.Response{Allow: true}
	}

	imagename := utils.GetImageFromContainer(cName)
	if imagename == "" {
		return authorization.Response{Msg: "Bad container name"}
	}

	userns, ok := utils.UsernsInContainer(cName)
	if ok && !userns {
		return authorization.Response{Allow: true}
	}

	user, ok := docker_options["User"].(string)
	if !ok || user == "" {
		return authorization.Response{Err: "User not specified. Use -u flag."}
	}

	if !utils.UsersEqual(user, req.User) {
		return authorization.Response{Err: "Not allowed for user " + user + ". Did you set group id?"}
	}

	if !utils.UserCreatedContainer(user, cName) {
		return authorization.Response{Msg: "user " + user +
			" is not allowed for this container "}
	}

	return authorization.Response{Allow: true}

}

func CommandInContainer(req authorization.Request, suffix string) (string, bool) {

	re := regexp.MustCompile("/containers/(.+)" + suffix)
	matches := re.FindStringSubmatch(req.RequestURI)
	if len(matches) == 2 {
		return matches[1], true
	}
	return "", false

}

func (p *novolume) AuthZReq(req authorization.Request) authorization.Response {

	if isRequestFromRoot(req) {
		return authorization.Response{Allow: true}
	}

	switch req.RequestMethod {
	case "POST":
		match, err := regexp.MatchString("/containers/create", req.RequestURI)
		if err == nil && match {
			return processStartContainer(req)
		}

		if cname, ok := CommandInContainer(req, "/exec"); ok {
			return processExecContainer(req, cname)
		}

		if cname, ok := CommandInContainer(req, "/"); ok {
			return processContainerCommand(req, cname)
		}

		return authorization.Response{Allow: true}

	case "GET":
		return authorization.Response{Allow: true}
	case "DELETE":
		if cname, ok := CommandInContainer(req, "\\?force"); ok {
			return processContainerCommand(req, cname)
		}
		if cname, ok := CommandInContainer(req, ""); ok {
			return processContainerCommand(req, cname)
		}
	}

	return authorization.Response{Msg: "Command allowed for root only "}
}

func (p *novolume) AuthZRes(req authorization.Request) authorization.Response {
	return authorization.Response{Allow: true}
}
