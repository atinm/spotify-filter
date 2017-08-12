@ECHO OFF

IF "%GOPATH%"=="" GOTO NOGO
IF NOT EXIST %GOPATH%\bin\2goarray.exe GOTO INSTALL
:POSTINSTALL
IF "%1"=="" GOTO NOICO
IF NOT EXIST %1 GOTO BADFILE
IF "%2"=="" GOTO NOVARIABLE
IF "%3"=="" GOTO NOFILENAME

ECHO Creating %3
ECHO //+build windows > %3
ECHO. >> %3
TYPE %1 | %GOPATH%\bin\2goarray Data icon >> %3
GOTO DONE

:CREATEFAIL
ECHO Unable to create output file
GOTO DONE

:INSTALL
ECHO Installing 2goarray...
go get github.com/cratonica/2goarray
IF ERRORLEVEL 1 GOTO GETFAIL
GOTO POSTINSTALL

:GETFAIL
ECHO Failure running go get github.com/cratonica/2goarray.  Ensure that go and git are in PATH
GOTO DONE

:NOGO
ECHO GOPATH environment variable not set
GOTO DONE

:NOICO
ECHO Please specify a .ico file
GOTO DONE

:NOVARIABLE
ECHO Please specify a Variable name
GOTO DONE

:NOFILENAME
ECHO Please specify an output file name
GOTO DONE

:BADFILE
ECHO %1 is not a valid file
GOTO DONE

:DONE

