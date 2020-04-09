#!/bin/bash

LOG=`date "+%Y-%m-%dT%H-%M-%S.%N"`
exec server.exe > ${LOG}.txt 2>&1
