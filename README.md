Container update checker
========================

Tool that checks if a running container's image and tag is matching the same image and tag on the remote repository.

There are a handful a situation where the image and tag of a container does not match that same image and tag on a docker registry:
- if a new image has been pushed with the same tag. for example `latest`
- if a new image has been built with the same tag locally after the container was started.

The tool will report such discrepancy.

```
 hugues in ~/go/src/github.com/hmalphettes/containertagchecker
⚡ DOCKER_HOST=http://localhost:2375 ~/go/bin/containertagchecker
CONTAINER ID                                                            REPO:TAG                     UPTODATE
728d8cc3002bfaee98742e116190d9044cf8439adb443de140b7dcb67dfe020a        jwilder/whoami:latest        TRUE
f679f16769ca287b9685c1059328d8c3d4cf8413ff8068f020851df6fd17e5a7        <none>:<none>                FALSE
96e1551174885d0761e6bfda4b9bfabe11c50070e11223dd8ddcde712a1ae04b        jwilder/whoami:latest        TRUE
13d0117952483229741d50d8cb6418bbb95f3ce76970e6a99aa5dd6d7ce3052c        <none>:<none>                FALSE
3b8310b9f5662c995d0113cd9031bdc212cdfe184a0ec67f436bc2c01e9b72d9        jwilder/whoami:hugues        NOTFOUND
```

![containertagchecker](containertagchecker.png)

Notes:
======

It is interesting to observe that the canonical way to address an image - the `Image's digest` - is not related to the way the docker images are addressed by the docker-engine: by `Image ID`.

This tool was coded in golang. We could also script it in bash.

