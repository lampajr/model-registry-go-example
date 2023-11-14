#!/bin/bash

docker rm $(docker ps -af name=mlmd-server -q) -f