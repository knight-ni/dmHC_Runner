rem amd64 arm64 ppc64le ppc64 loong64//pkg\gopsutil\host\host_linux.go:98:22: undefined: sizeOfUtmp
set archlist=amd64 arm64 ppc64le
set oslist=linux
setlocal enableDelayedExpansion

for %%a in (%archlist%) do (
   for  %%b in (%oslist%) do (
    set GOARCH=%%a
    set GOOS=%%b
      D:\Programes\GoLANG1.20.3\go1.20.3\bin\go build -ldflags "-s -w" -o bin/%%b/dmHC_Runner_%%b_%%a
    )
)
pause