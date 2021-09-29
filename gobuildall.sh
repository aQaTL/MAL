#!/bin/bash

mkdir -p build/windows_x64 \
	build/windows_x86 \
	build/linux_x64 \
	build/linux_x86 \
	build/darwin_x64 \
	build/darwin_x86 \
	build/darwin_arm64

rootDir=`pwd`

if [ "$(uname -s)" = "Windows" ]; then
	GO_EXE=go.exe
else
	GO_EXE=go
fi


echo Building for Linux amd64
cd $rootDir/build/linux_x64
GOOS=linux GOARCH=amd64 $GO_EXE build ../.. &

echo Building for Linux 386
cd $rootDir/build/linux_x86
GOOS=linux GOARCH=386 GO386=softfloat $GO_EXE build ../.. &

echo Building for Windows amd64
cd $rootDir/build/windows_x64
GOOS=windows GOARCH=amd64 $GO_EXE build ../.. &

echo Building for Windows 386
cd $rootDir/build/windows_x86
GOOS=windows GOARCH=386 GO386=softfloat $GO_EXE build ../.. &

echo Building for Darwin amd64
cd $rootDir/build/darwin_x64
GOOS=darwin GOARCH=amd64 $GO_EXE build ../.. &

echo Building for Darwin 386
cd $rootDir/build/darwin_x86
GOOS=darwin GOARCH=386 GO386=softfloat $GO_EXE build ../.. &

echo Building for Darwin arm64
cd $rootDir/build/darwin_arm64
GOOS=darwin GOARCH=arm64 $GO_EXE build ../.. &

wait

echo Done building

cd $rootDir/build

for dir in */; do
	dir=${dir:0:(-1)}
	echo "Compressing $dir" 
	7z a -t7z -m0=lzma -mx=9 -mfb=64 -md=32m -ms=on $dir.7z ./$dir/* > /dev/null &
	7z a -mx=9 -mfb=64 $dir.zip ./$dir/* > /dev/null &
done

wait

echo Done compressing

cd $rootDir
