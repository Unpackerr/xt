xt(1) -- eXtractor Tool - Recursively decompress archives
===

SYNOPSIS
---

`xt [options] [path [path] [path] ...]`  
`xt --job-file /tmp/job1 -j /tmp/job2`\

DESCRIPTION
---

*   This application recursively extracts compressed archives.
*   Provide directories to search or provide files to extract.
*   Supports: ZIP, RAR, GZIP, BZIP2, TAR, TGZ, TBZ2, 7ZIP, ISO9660
*   Supports: Z, AR, BR, CPIO, DEB, LZ/4, LZIP, LZMA2, S2, SNAPPY
*   Supports: RPM, SZ, TLZ, TXZ, ZLIB, ZSTD, BROTLI, ZZ
*   ie: *.zip *.rar *.r00 *.gz *.bz2 *.tar *.tgz *.tbz2 *.7z *.iso (and others)

OPTIONS
---

`xt [-o </dir>] [-d <#>] [-m <#>] [-e <.ext>] [-P <p4ss,words>] [-p] [paths]`

-o _directory_, --output _directory_  
    Provide a file system _directory_ where content is written.
    The default output _directory_ is the current working directory.

-S, --squash-root
    If an archive contains only a single folder in the root directory,
    then the contents of that folder are moved into the output folder.
    The now-empty original root folder is deleted.

-d _count_, --max-depth _count_  
    This option limits how deep into the file system xt recurses.
    The default is (0) unlimited. Setting to 1 disables recursion.

-m _count_, --min-depth _count_  
    This option determines if archives should only be found deeper
    into the file system. The default is (0) root. Archives are only
    extracted from _count_ sub directories deep or deeper.

-P _password_, --password _password_  
    Provided _passwords_ are attempted against extraction of encrypted
    rar and/or 7zip archives. The `-p` option may be provided many times.

-e _.ext_, --extension _.ext_  
    Only extract archives with these extensions. Include the leading dot.
    The `-e` option may be provided many times. 
    Use `-v` for supported extensions. <- Your input must match the
    supported extensions. Unknown extensions are still ignored.

-j _file_, --job-file _file_  
    The options above create a single job. If you want more control,
    you may provide one or more job files. Each _file_ may define the
    input, output, depths and passwords, etc. Acceptable formats are
    xml, json, toml and yaml. TOML is the default. See JOB FILES below.

-p, --preserve-paths
    This option determines if the archives will be extracted to their
    parent folder. Using this flag will override the --output option.

-V, --verbose
    Verbose logging prints the extracted file paths.

-D, --debug
    Enable debug output.

-v, --version  
    Display version and exit.

-h, --help  
    Display usage and exit.

JOB FILES
---

If  `include_suffix` is provided `exclude_suffix` is ignored.

Example TOML job file:

    paths     = [ '/path1', '/another/path' ]
    output    = '.'
    passwords = [ 'password1', '''password"With'Specials!''', 'pass3']
    exclude_suffix = ['.iso', '.gz']
    include_suffix = ['.zip', '.rar', '.r00']
    max_depth = 0
    min_depth = 1
    file_mode = 644
    dir_mode  = 755
    verbose   = false
    debug     = false
    preserve_paths = false

AUTHOR
---

*   David Newhall II - 1/20/2024

LOCATION
---

*   https://unpackerr.zip/xt
*   https://golift.io/gpg
*   /usr/bin/xt || /usr/local/bin/xt
