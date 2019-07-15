#!/bin/bash

docker stop beautiful-jekyll
docker run --rm -d -p 0.0.0.0:4000:4000 --name beautiful-jekyll -v "$PWD":/srv/jekyll anjiawei/beautiful-jekyll
