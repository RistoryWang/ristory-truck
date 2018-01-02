#!/bin/sh

tarfile="ristory-truck-$1.tar.gz"

echo "开始打包$tarfile..."

export GOARCH=amd64
export GOOS=linux

bee pack -exs=".go:.DS_Store:.tmp:.log" -exr=data

mv ristory-truck.tar.gz $tarfile
#mv ristory-truck-$1.tar.gz $HOME/Desktop/128-upload.app/

ftp -n <<- EOF
open 192.168.10.168
user ftp xxxxxxxxxx
binary
cd /
put $tarfile
bye
EOF

rm -rf $tarfile
