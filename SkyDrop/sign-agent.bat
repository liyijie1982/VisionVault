set path=%path%;C:\Program Files\Microsoft SDKs\Windows\v6.0A\bin;
signtool sign /i globalsign /t http://timestamp.digicert.com SyncAgent-install/SyncAgent.exe
signtool sign /i globalsign /t http://timestamp.digicert.com SyncAgent-install/SyncAgentAutoUpdate.exe

@echo finish!
