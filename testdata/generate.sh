#!/bin/zsh

protoc -I=. --go_out=paths=source_relative:. --go-vtproto_out=paths=source_relative,features=marshal+unmarshal+size:. message.proto
