#!/bin/sh
set -xe

./migrate up
exec ./kypidbot
