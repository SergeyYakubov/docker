package utils

import (
	//	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	//	"utils"
)

const testdigest = "sha256:d26de97585f520712e9936074588df7df9f8c27e2e08d29db55affb612be7a93"
const testimage = "centos_mpi"
const testname = "test"

var repotests = []struct {
	name             string
	repo, image, tag string
}{
	{"test1/test2:test3", "test1", "test2", "test3"},
	{" test1/test2:test3 ", "test1", "test2", "test3"},
	{"test1/test2", "test1", "test2", "latest"},
	{"test1", "", "", ""},
}

func TestSplitImageName(t *testing.T) {
	for _, test := range repotests {
		repo, image, tag := SplitImageName(test.name)
		assert.Equal(t, test.image, image, "Error getting image from "+test.name)
		assert.Equal(t, test.repo, repo, "Error getting repo from "+test.name)
		assert.Equal(t, test.tag, tag, "Error getting tag from "+test.name)
	}
}

var repodigests = []struct {
	in  string
	out string
}{
	{systemsRegistry + "/" + testimage, testdigest},
	{"   " + systemsRegistry + "/" + testimage + "   ", testdigest},
	{"bla-bla", ""},
}

func TestRepoDigist(t *testing.T) {
	for _, repodigest := range repodigests {
		assert.Equal(t, repodigest.out, GetRepoDigist(repodigest.in), "Error comparing repo digest")
	}
}

func TestLocalDigist(t *testing.T) {
	for _, repodigest := range repodigests {
		assert.Equal(t, GetLocalDigist(repodigest.in), repodigest.out, "Error comparing local digest")
	}
}

var images = []struct {
	in  string
	out bool
}{
	{systemsRegistry + "/" + testimage, true},
	{"  " + systemsRegistry + "/" + testimage + "  ", true},
	{"bla-bla", false},
	{systemsRegistry + "/blabla", false},
}

func TestVerifiedImage(t *testing.T) {
	for _, image := range images {
		assert.Equal(t, VerifiedImage(image.in), image.out, "Error verifying image for "+image.in)
	}
}

var containers = []struct {
	name string
	user string
	out  bool
}{
	{testname, "yakubov", true},
	{"blabla", "yakubov", false},
	{"blabla", "yakubo", false},
}

func TestContainers(t *testing.T) {
	for _, container := range containers {
		err := UserCreatedContainer(container.user, container.name)
		val := err != nil
		assert.Equal(t, val, container.out, "Error verifying container for "+container.name)
	}
}

var userdatas = []struct {
	fname string
	uid   string
	gid   string
}{
	{slurmdir + "/" + testname, "26655", "1000"},
	{slurmdir + "/" + testname, "26655", "26655"},
	{slurmdir + "/" + testname, "1000", "26655"},
	{"bla", "bla", "bla"},
}

func TestReadUserData(t *testing.T) {
	data := userdatas[0]
	uid, gid, err := ReadUserData(data.fname)
	if assert.Nil(t, err) {
		assert.Equal(t, data.gid, gid, "Error verifying gid for "+data.fname)
		assert.Equal(t, data.uid, uid, "Error verifying uid for "+data.fname)
	}

	data = userdatas[1]
	uid, gid, err = ReadUserData(data.fname)
	if assert.Nil(t, err) {
		assert.Equal(t, data.uid, uid, "Error verifying gid for "+data.fname)
		assert.NotEqual(t, data.gid, gid, "Error verifying uid for "+data.fname)
	}

	data = userdatas[2]
	uid, gid, err = ReadUserData(data.fname)
	if assert.Nil(t, err) {
		assert.NotEqual(t, data.uid, uid, "Error verifying gid for "+data.fname)
		assert.NotEqual(t, data.gid, gid, "Error verifying uid for "+data.fname)
	}

	data = userdatas[3]
	uid, gid, err = ReadUserData(data.fname)
	assert.NotNil(t, err)

}
