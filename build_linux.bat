rem amd64 arm64 ppc64le ppc64 loong64//pkg\gopsutil\host\host_linux.go:98:22: undefined: sizeOfUtmp
set archlist=amd64 arm64 ppc64le
set oslist=linux
setlocal enableDelayedExpansion

for %%a in (%archlist%) do (
   for  %%b in (%oslist%) do (
    set GOARCH=%%a
    set GOOS=%%b
      D:\gosdk\pkg\mod\golang.org\toolchain@v0.0.1-go1.22.5.windows-amd64\bin\go build -ldflags "-s -w" -pgo=auto -o bin/%%b/dmHC_Runner_%%b_%%a
    )
)
pause