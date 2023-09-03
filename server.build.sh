#!/bin/bash

REPO_NAME="https://github.com/senthilsweb/goduck.git"
OS="$(uname -s)"
GIT_COMMIT=$(git rev-parse --short HEAD)
TAG=$(git describe --exact-match --abbrev=0 --tags ${COMMIT} 2> /dev/null || true)
TAG=${TAG:="prod"}
DATE=$(date +'%Y-%m-%d')
APP_NAME='goduck'
BIN_NAME=$APP_NAME'_'
DIST='./release/'
DEST='./.temp/'
INSTALL_BUNDLE=$APP_NAME'.tar.gz'
echo $INSTALL_BUNDLE
echo 'Operating System = ['$OS']'

echo "Building binaries"
echo Git commit: $GIT_COMMIT Version: $TAG Build date: $DATE


go generate

# MAC
export GOARCH="amd64"
export GOOS="darwin"
export CGO_ENABLED=1
go build -ldflags "-X $REPO_NAME/cmd.GitCommit=$GIT_COMMIT -X $REPO_NAME/cmd.Version=$TAG -X $REPO_NAME/cmd.BuildDate=$DATE" -o $DEST$BIN_NAME'mac_amd64' -v .
#GOOS=darwin GOARCH=amd64 go build

#LINUX
export GOARCH="amd64"
export GOOS="linux"
export CGO_ENABLED=1
go build -ldflags "-X $REPO_NAME/cmd.GitCommit=$GIT_COMMIT -X $REPO_NAME/cmd.Version=$TAG -X $REPO_NAME/cmd.BuildDate=$DATE" -o $DEST$BIN_NAME'linux_amd64' -v


export GOARCH="386"
export GOOS="linux"
export CGO_ENABLED=1
go build -ldflags "-X $REPO_NAME/cmd.GitCommit=$GIT_COMMIT -X $REPO_NAME/cmd.Version=$TAG -X $REPO_NAME/cmd.BuildDate=$DATE" -o $DEST$BIN_NAME'linux_386' -v

#WINDOWS
#export GOARCH="386"
#export GOOS="windows"
#export CGO_ENABLED=1
#go build -ldflags "-X $REPO_NAME/cmd.GitCommit=$GIT_COMMIT -X $REPO_NAME/cmd.Version=$TAG -X $REPO_NAME/cmd.BuildDate=$DATE" -o $DEST$BIN_NAME'windows_386.exe' -v

#export GOARCH="amd64"
#export GOOS="windows"
#export CGO_ENABLED=1
#go build -ldflags "-X $REPO_NAME/cmd.GitCommit=$GIT_COMMIT -X $REPO_NAME/cmd.Version=$TAG -X $REPO_NAME/cmd.BuildDate=$DATE" -o $DEST$BIN_NAME'windows_amd64.exe' -v


cp server.service.txt $DEST$APP_NAME'.service'
cp server.install.md $DEST$APP_NAME'.install.md'
cp server.install.sh $DEST'server.install.sh'

if [ "$(uname)" == "Darwin" ]; then
    #For Mac
    sed -i "" 's/{{app}}/'$APP_NAME'/g' $DEST$APP_NAME'.service'
    sed -i "" 's/{{app}}/'$APP_NAME'/g' $DEST'server.install.sh'
    sed -i "" 's/{{app}}/'$APP_NAME'/g' $DEST$APP_NAME'.install.md'
    sed -i "" 's/{{tar_file}}/'$INSTALL_BUNDLE'/g' $DEST$APP_NAME'.install.md'
else
    #For Linux
    sed -i 's/{{app}}/'$APP_NAME'/g' $DEST$APP_NAME'.service'
    sed -i 's/{{app}}/'$APP_NAME'/g' $DEST'server.install.sh'
    sed -i 's/{{app}}/'$APP_NAME'/g' $DEST$APP_NAME'.install.md'
    sed -i 's/{{tar_file}}/'$INSTALL_BUNDLE'/g' $DEST$APP_NAME'.install.md'
fi
echo $DIST$INSTALL_BUNDLE
tar -czvf $DEST$INSTALL_BUNDLE -C $DIST .
#rm -rf $DEST
echo "Build complete"