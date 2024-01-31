xt(1) -- eXtractor Tool - Recursively decompress archives
===

SYNOPSIS
---

`xt [options] [path [path] [path] ...]`

DESCRIPTION
---

*   This application recursively extracts compressed archives.
*   Provide directories to search or provide files to extract.
*   Supports ZIP, RAR, GZIP, BZIP2, TAR, TGZ, TBZ2, 7ZIP, ISO9660
*   ie. *.zip *.rar *.r00 *.gz *.bz2 *.tar *.tgz *.tbz2 *.7z *.iso

OPTIONS
---

`xt [-o </path>] [-d <#>] [-m <#>] [-p <p4ss,words>] [paths]`

    -o, --output </path>
        Provide a file system path where content should be written.
        The default output path is the current working directory.

    -d, --max-depth <child count>
        This option limits how deep into the file system xt recurses.
        The default is (0) unlimited. Setting to 1 disables recursion.

    -m, --min-depth <child count>
        This option determines if archives should only be found deeper
        into the file system. The default is (0) root. Archives are only
        extracted from <child count> sub directories deep or deeper.

    -P, --password <p4ss word>,<pass w0rd>
        Provided password(s) are attempted against extraction of encrypted
        rar and/or 7zip archives. The -p option may be provided many times.

    -j, --job-file <job file>
        The options above create a single job. If you want more control,
        you may provide one or more job files. Each file may define the
        input, output, depths and passwords, etc. Acceptable formats are
        xml, json, toml and yaml. TOML is the default. See JOB FILES below.

    -v, --version
        Display version and exit.

    -h, --help
        Display usage and exit.

JOB FILES
---

Example TOML job file:

    paths     = [ '/path1', '/another/path' ]
    output    = '.'
    passwords = [ 'password1', '''password"With'Specials!''', 'pass3']
    exclude_suffix = ['.iso', '.gz']
    max_depth = 0
    min_depth = 1
    file_mode = 644
    dir_mode  = 755

AUTHOR
---

*   David Newhall II - 1/20/2024

LOCATION
---

*   https://unpackerr.zip/xt
*   https://golift.io/gpg
*   /usr/bin/xt || /usr/local/bin/xt
