# gzip file scanner

This scans for gzip files in a file or disk image, and can do so with multiple threads (useful if source is SSD/NVMe).

Usage:

	gzscan /dev/sda

Note that it is quite likely that any large gzip file found this way will not be readable, as other files may be written in the middle of this file. Finding the header however greatly helps recovering the rest of the file.

## output

When a likely gzip file is found, it is shown as follows:

	2022/04/21 12:21:09 found likely gzip: pos=934375424 stamp=2019-03-23 00:00:01 +0900 JST os=Unix filename=xxxxx.log

False positives will also appear, however if for example the filename is shown, this is very likely a good match.

Gzip will not store the filename if the input was piped (ie. `getlog | gzip >file.gz`), in this case you may try to decode the file manually.
