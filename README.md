# mfrep
MFREP (multi-file replace) is a cli tool for quickly replacing the parts of specific files

DESCRIPTION:
Iterates through the current working directory contents and marks the files which will be affected by the replace. If the --path option is specified, only the files with suitable name will be marked.

In each file finds all substring, that satisfy <search_regexp> and replaces them with <replace>.

Notice that the <replace> argument can contain default Sprintf placeholders (%v or %s) for submatches of regexp. The number of capturing groups in regex should be equal to the number of placeholders.

For recursive iteration use --recursive option.