#!/bin/bash

cd web

echo "Building assets"
yarn build > /dev/null

cd ./..

echo "Compiling binary"
packr build

echo "Doing cleanup"
rm -rf ./web/dist
