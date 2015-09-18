#!/bin/bash
docker build -t composer22/coreos-deploy_build .
docker run -v /var/run/docker.sock:/var/run/docker.sock -v $(which docker):$(which docker) -ti --name coreos-deploy_build composer22/coreos-deploy_build
docker rm coreos-deploy_build
docker rmi composer22/coreos-deploy_build
