# gzip file scanner

This scans for gzip files in a file or disk image, and can do so with multiple threads (useful if source is SSD/NVMe).

Usage:

	gzscan /dev/sda

Note that it is quite likely that any large gzip file found this way will not be readable, as other files may be written in the middle of this file. Finding the header however greatly helps recovering the rest of the file.
