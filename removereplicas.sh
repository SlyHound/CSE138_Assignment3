#!/bin/bash

docker stop replica1 && docker stop replica2 && docker stop replica3

docker rm replica1 && docker rm replica2 && docker rm replica3

docker network rm mynet

docker network create --subnet=10.10.0.0/16 mynet

docker build -t assignment3-img .