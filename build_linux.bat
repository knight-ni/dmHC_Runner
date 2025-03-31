rem amd64 arm64 ppc64le ppc64 loong64
set archlist=amd64 arm64 ppc64le
set oslist=linux
set currentPath=%~dp0
set CGO_ENABLED=0

set year=%date:~0,4%
set month=%date:~5,2%
set day=%date:~8,2%
set hour=%time:~0,2%
set min=%time:~3,2%
set sec=%time:~6,2%

set "first=%hour:~0,1%"
if "%first%"==" " set "hour=0%hour:~1%"

set "first=%min:~0,1%"
if "%first%"==" " set "min=0%min:~1%"

set "first=%sec:~0,1%"
if "%first%"==" " set "sec=0%sec:~1%"

set tm=%year%%month%%day%%hour%%min%%sec%
setlocal enableDelayedExpansion

cd src
for %%a in (%archlist%) do (
   for  %%b in (%oslist%) do (
    set GOARCH=%%a
    set GOOS=%%b
      D:\gosdk\go1.23.0\bin\go build -ldflags "-s -w --extldflags '-static -fpic' -X main.buildTime=%tm% " -pgo=auto -o ..\bin\%%b\dmHC_Runner_%%b_%%a
	  MOVE %currentPath%\bin\%%b\dmHC_Runner_%%b_%%a %currentPath%dmhc\
    )
)
cd ..
