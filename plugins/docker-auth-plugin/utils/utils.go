package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/user"
	"strings"

	"os"

	docker "github.com/fsouza/go-dockerclient"
)

const systemsRegistry = "max-adm01:5000"
const slurmdir = "/var/run/slurm"

func ImageFromSystemRepo(imagename string) bool {
	registry, _, _ := SplitImageName(imagename)
	if registry != systemsRegistry {
		return false
	}
	return true
}

func VerifiedImage(imagename string) bool {

	if !ImageFromSystemRepo(imagename) {
		return false
	}

	repodig := GetRepoDigist(imagename)
	localdig := GetLocalDigist(imagename)
	if localdig == "" || repodig == "" || localdig != repodig {
		return false
	}

	return true

}

func UserGroup(username, usergroup string) bool {

	u, err := GetUser(username)
	if err != nil {
		return false
	}

	// ugly,I know  - allow afs group
	if len(usergroup) == 10 && strings.HasPrefix(usergroup, "109") {
		return true
	}

	g, err := user.LookupGroup(usergroup)
	if err != nil {
		g, err = user.LookupGroupId(usergroup)
		if err != nil {
			return false
		}
	}

	groups, err := u.GroupIds()

	if err != nil {
		return false
	}
	for _, group := range groups {
		if group == g.Gid {
			return true
		}
	}

	return false

}

func UsersEqual(user, certUser string) bool {

	u, err := GetUser(user)
	if err != nil {
		return false
	}

	uc, err := GetUser(certUser)
	if err != nil {
		return false
	}

	if *u != *uc {
		return false
	}

	return true

}

func GetUser(str string) (u *user.User, err error) {
	str = strings.TrimSpace(str)
	split := strings.SplitN(str, ":", 2)
	var username, usergroup string
	if len(split) == 2 {
		username = split[0]
		usergroup = split[1]
	} else if len(split) == 1 {
		username = split[0]
		usergroup = ""
	} else {
		err = errors.New("Cannot extract user")
		return
	}

	u, err = user.Lookup(username)
	if err != nil {
		u, err = user.LookupId(username)
		if err != nil {
			return
		}
	}

	if usergroup == "" {
		err = errors.New("Group not set")
		return
	}

	g, err := user.LookupGroup(usergroup)
	if err != nil {
		g, err = user.LookupGroupId(usergroup)
		if err != nil {
			return
		}
	}

	groups, err := u.GroupIds()
	if err != nil {
		return
	}
	for _, group := range groups {
		if group == g.Gid {
			return
		}
	}

	return u, errors.New("Group not found")

}

func SplitImageName(str string) (repo, image, tag string) {

	str = strings.TrimSpace(str)
	split := strings.SplitN(str, "/", 2)
	if len(split) == 2 {
		repo = split[0]
	} else {
		return
	}

	split = strings.SplitN(split[1], ":", 2)
	if len(split) > 0 {
		image = split[0]
	}
	if len(split) > 1 {
		tag = split[1]
	} else {
		tag = "latest"
	}

	return
}

func GetContainer(cName string) (*docker.Container, error) {
	cName = strings.TrimSpace(cName)

	client := createClient()

	return client.InspectContainer(cName)

}

func GetContainerImageName(name string) (*docker.Image, error) {
	name = strings.TrimSpace(name)

	client := createClient()

	cont, err := client.InspectContainer(name)
	if err != nil {
		return nil, err
	}

	return client.InspectImage(cont.Image)

}

func GetImageFromContainer(cName string) string {

	image, err := GetContainerImageName(cName)

	if err != nil {
		return ""
	}

	if len(image.RepoTags) < 1 {
		return ""
	}
	return image.RepoTags[0]
}

func UsernsInContainer(cName string) (bool, bool) {

	cont, err := GetContainer(cName)
	if err != nil {
		return true, false
	}
	config := cont.HostConfig
	if config == nil {
		return true, false
	}

	return config.UsernsMode == "host", true
}

func createClient() *docker.Client {
	endpoint := "tcp://localhost:2376"
	path := "/home/yakubov/.docker"
	ca := fmt.Sprintf("%s/ca.pem", path)
	cert := fmt.Sprintf("%s/cert.pem", path)
	key := fmt.Sprintf("%s/key.pem", path)
	client, _ := docker.NewTLSClient(endpoint, cert, key, ca)
	return client
}

func GetLocalDigist(imagename string) string {

	imagename = strings.TrimSpace(imagename)

	client := createClient()
	image, err := client.InspectImage(imagename)
	if err != nil {
		return ""
	}

	if len(image.RepoDigests) == 0 {
		return ""
	}

	dig_str := image.RepoDigests[0]

	tmp := strings.Split(dig_str, "@")
	if len(tmp) < 2 {
		return ""
	}

	return tmp[1]
}

func GetRepoDigist(imagename string) string {

	registry, image, tag := SplitImageName(imagename)
	addr := "http://" + registry + "/v2/" + image + "/manifests/" + tag
	client := &http.Client{}
	req, err := http.NewRequest("HEAD", addr, nil)

	if err != nil {
		return ""
	}

	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	resp, err := client.Do(req)

	if err != nil {
		return ""
	}

	digest, ok := resp.Header["Docker-Content-Digest"]
	if ok {
		return digest[0]
	} else {
		return ""
	}
}

func ReadUserData(filename string) (uid, gid string, err error) {
	buff, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	s := strings.Fields(string(buff[:]))
	if len(s) < 2 {
		err = errors.New("cannot read uid,gid from " + filename)
		return
	}
	uid, gid = s[0], s[1]
	return
}

func GetContainerUser(container string) (string, error) {
	cont, err := GetContainer(container)
	if err != nil {
		return "", err
	}

	hostConfig := cont.HostConfig
	if hostConfig == nil {
		return "", errors.New("Cannot inspect container")
	}

	if hostConfig.UsernsMode != "host" {
		return "", nil
	}

	config := cont.Config

	if config == nil {
		return "", errors.New("Cannot inspect container")
	}

	return config.User, nil

}

func UserCreatedContainer(username, container string) bool {

	containerUser, err := GetContainerUser(container)
	if err != nil {
		return false
	}

	if containerUser == "" {
		return true
	}
	return UsersEqual(containerUser, username)

}

func CheckBind(str string) error {
	split := strings.Split(str, ":")
	if len(split) != 2 {
		return nil
	}

	hostdir := split[0]

	if !strings.HasPrefix(hostdir, "/") {
		return nil
	}

	if _, err := os.Stat(hostdir); os.IsNotExist(err) {
		return errors.New("Path does not exist: " + hostdir)
	}

	return nil
}
