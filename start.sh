#!/bin/bash

sudo docker stop beautiful-jekyll
sudo docker run --rm -d -p 4000:4000 --name beautiful-jekyll -v "$PWD":/srv/jekyll anjiawei/beautiful-jekyll
