package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
	"text/tabwriter"
)

// repository -> jwtToken
var jwtTokens = make(map[string]string)

func main() {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	imageIdToRepoTag, imageIdToImageDigest, err := GetLocalImageDigestsByRepoTag(cli)

	// imageId -> remoteImageDigest
	var remoteImageDigests = make(map[string]string)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 8, ' ', 0)
	fmt.Fprintln(w, "CONTAINER ID\tREPO:TAG\tUPTODATE")
	//	fmt.Fprintln(w, "CONTAINER ID\tREPO:TAG\tUPTODATE\tLOCAL DIGEST\tREMOTE DIGEST")
	for _, container := range containers {
		repoTag := imageIdToRepoTag[container.ImageID]
		localImageDigest := imageIdToImageDigest[container.ImageID]
		if repoTag == "<none>:<none>" {
			fmt.Fprintln(w, fmt.Sprintf("%s\t%s\t%s", container.ID, repoTag, "FALSE"))
			//			fmt.Fprintln(w, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", container.ID, repoTag, "FALSE", localImageDigest, ""))
			continue
		}

		remoteImageDigest, err := GetImageDigestFromRegistry(repoTag, remoteImageDigests)
		if err != nil {
			panic(err)
		}
		upToDate := "TRUE"
		if remoteImageDigest != localImageDigest {
			upToDate = "FALSE"
		} else if remoteImageDigest == "" {
			upToDate = "NOTFOUND"
		}
		fmt.Fprintln(w, fmt.Sprintf("%s\t%s\t%s", container.ID, repoTag, upToDate))
		//		fmt.Fprintln(w, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", container.ID, repoTag, upToDate, localImageDigest, remoteImageDigest))
	}
	w.Flush()

}

func GetLocalImageDigestsByRepoTag(cli *client.Client) (map[string]string, map[string]string, error) {
	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		return nil, nil, err
	}

	imageIdToRepoTag := make(map[string]string)
	imageIdToImageDigest := make(map[string]string)
	for _, image := range images {
		if len(image.RepoTags) != 0 {
			//			fmt.Printf("HERE IS A LOCAL IMAGE %v\n", image)
			//fmt.Printf("%v\t%s\t%v\n", image.RepoTags, image.ID, image.RepoDigests)
			imageIdToRepoTag[image.ID] = image.RepoTags[0]
		}
		for _, imgdigest := range image.RepoDigests {
			repoAndDigest := strings.Split(imgdigest, "@")
			imageIdToImageDigest[image.ID] = repoAndDigest[1]
			if len(image.RepoTags) == 0 {
				// When an image has been overridden locally by a newly built image with the same tag
				// we end-up with no repoTag at all.
				// we can still get the repository from the repoDigest:
				imageIdToRepoTag[image.ID] = repoAndDigest[0] + ":<none>"
			}
		}
	}
	return imageIdToRepoTag, imageIdToImageDigest, nil
}

/**
 * Getting an image digest from the docker repository.
 * Reference: https://stackoverflow.com/a/41830007/1273401
 *
 * Note: we could also use https://godoc.org/docker.io/go-docker#Client.DistributionInspect
 */
func GetImageDigestFromRegistry(repoTag string, cache map[string]string) (string, error) {
	toks := strings.Split(repoTag, ":")
	repository := toks[0]
	if _, ok := cache[repository]; ok {
		return cache[repository], nil
	}
	var tag string
	if len(toks) > 1 {
		tag = toks[1]
	} else {
		tag = "latest"
	}
	if tag == "<none>" {
		return "", nil
	}
	jwtToken, err := GetJWTTokenFromRegistry(toks[0])
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("HEAD", "https://index.docker.io/v2/"+repository+"/manifests/"+tag, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return resp.Header.Get("Docker-Content-Digest"), nil
}

func GetJWTTokenFromRegistry(repository string) (string, error) {
	if _, ok := jwtTokens[repository]; ok {
		return jwtTokens[repository], nil
	}
	resp, err := http.Get("https://auth.docker.io/token?service=registry.docker.io&scope=repository:" + repository + ":pull")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var body struct {
		Token string `json:"token"`
	}
	json.NewDecoder(resp.Body).Decode(&body)

	return body.Token, nil
}
