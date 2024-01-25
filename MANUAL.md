xt(1) -- eXtractor Tool - Recursively decompress archives
===

SYNOPSIS
---

`xt [options] [path [path] [path] ...]`

DESCRIPTION
---

*   This application recursively extracts compressed archives.
    Provide directories to search or provide files to extract.

*   Supports ZIP, RAR, GZIP, BZIP2, TAR, TGZ, TBZ2, 7ZIP, ISO9660
    ie. *.zip *.rar *.r00 *.gz *.bz2 *.tar *.tgz *.tbz2 *.7z *.iso

OPTIONS
---

    -o, --output <path>
        Provide a file system path where content should be written.
        The default output path is the current working directory.

    -v, --version
        Display version and exit.

    -h, --help
        Display usage and exit.


AUTHOR
---

*   David Newhall II - 1/20/2024

LOCATION
---

*   https://unpackerr.zip/xt
*   https://golift.io/gpg
*   /usr/bin/xt || /usr/local/bin/xt
