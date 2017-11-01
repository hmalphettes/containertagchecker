package main

import (
	"fmt"
	"github.com/docker/docker/client"
	"strings"
	"testing"
)

func TestGetLocalImageDigestsByRepoTag(t *testing.T) {
	cli, e := client.NewEnvClient()
	if e != nil {
		panic(e)
	}
	imageIdToRepoTag, imageIdToImageDigest, err := GetLocalImageDigestsByRepoTag(cli)
	if err != nil {
		t.Error("Not expecting an error %v", err)
	}
	if imageIdToRepoTag == nil {
		t.Error("Could not find any images? %v", imageIdToRepoTag)
	}
	if imageIdToImageDigest == nil {
		t.Error("Could not find any images? %v", imageIdToRepoTag)
	}
}

func TestJWTToken(t *testing.T) {
	token, err := GetJWTTokenFromRegistry("jwilder/whoami")
	if err != nil {
		t.Error("Not expecting an error %v", err)
	} else {
		fmt.Println("JWT Token: " + token)
	}
}

func TestGetImageDigestFromRegistry(t *testing.T) {
	digest, err := GetImageDigestFromRegistry("jwilder/whoami:latest", make(map[string]string))
	if err != nil {
		t.Error("Not expecting an error %v", err)
		return
	}
	strings.HasPrefix(digest, "sha256:")
	fmt.Println("Digest of jwilder/whoami:latest " + digest)
}
