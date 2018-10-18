grepy
=====


This project implements a simple grep like program in multiple 
languages. The program needs to get as input:
* regex
* list of files, can be empty or contain only '-'
* mutually exclusive optional parameters for formatting

The program will look for the pattern in the files.  For each 
line that matches the program will print the filename, 
line number and the line formatted in one of the format options

* -c / --color: highlight the matching part in color
* -u / --underline: mark the match with '^' chars below it
* -m / --machine: (default) print in machine format - [file name]:[line number]:[match text]
