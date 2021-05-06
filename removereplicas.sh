#!/bin/bash

docker stop replica1 && docker stop replica2 && docker stop replica3

docker rm replica1 && docker rm replica2 && docker rm replica3