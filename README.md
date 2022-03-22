% BEHEAD(1) behead 1.0.0
% Bharath Chandra Sudheer
% March 2022

# NAME

**behead** - take all contents after skipping a number of lines.

# SYNOPSIS

**behead** \[**-n** *count*\] \[**-f** *INPUT-FILE*\] \[**-o** *OUTPUT-FILE*\]

# DESCRIPTION

This utility will display all content after the first *count* lines from the specified input file, or of the standard
input if no files are specified. The utility will write to the output file, or standard out if no file is specified.
If the utility detects a specified output file, it will report progress to the standard output.

The following options are available:

**-n** *count*
    Skip the first *count* lines of the input

**-f** *INPUT-FILE*
    Input will be read from this file, or standard in if not specified.

**-o** *OUTPUT-FILE*
    Output will be written to this file, or to standard out if not specified.

# EXIT STATUS

Returns 0 on success and >0 if an error occurs.

# EXAMPLES

To display the 42nd line onwards of the file *foo*, any of these can be used:

    $ behead -n 42 -f foo

To direct the output to a file *bar*:

    $ behead -n 42 -f foo -o bar

As reading from standard in and out can be done, there are many ways of doing the same:

    $ behead -n 42 < foo > bar
    $ behead -n 42 -o bar < foo
    $ cat foo | behead -n 42 > bar

# BUGS

If you find anything, please report it here http://github.com/bharathcs/behead/issues

# COPYRIGHT

Copyright © 2022 Bharath Chandra Sudheer. All rights reserved.

This work is licensed under the terms of the MIT license.  
For a copy, see https://opensource.org/licenses/MIT