rem amd64 arm64 ppc64le ppc64 loong64//pkg\gopsutil\host\host_linux.go:98:22: undefined: sizeOfUtmp
set archlist=amd64 arm64
set oslist=windows
setlocal enableDelayedExpansion

for %%a in (%archlist%) do (
   for  %%b in (%oslist%) do (
    set GOARCH=%%a
    set GOOS=%%b
      D:\gosdk\go1.21.5\bin\go build -ldflags "-s -w" -pgo=auto -o bin/%%b/dmHC_Runner_%%b_%%a.exe
   )
)
pause
