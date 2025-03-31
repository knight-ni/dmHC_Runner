set name=dmHC
set year=%date:~0,4%
set month=%date:~5,2%
set day=%date:~8,2%
set date=%year%%month%%day%
set currentPath=%~dp0
set zip="C:\Program Files\7-Zip\7z.exe"

%zip% a -tzip %currentPath%%name%_%date%.zip %currentPath%exe\*