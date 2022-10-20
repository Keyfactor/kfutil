#!/usr/bin/env bash
export PROJECTS_PATH=$HOME/GolandProjects
rm -rf vendor
go mod vendor
rm -rf vendor/github.com/Keyfactor/keyfactor-go-client
ln -s "$PROJECTS_PATH/keyfactor-go-client" vendor/github.com/Keyfactor/keyfactor-go-client