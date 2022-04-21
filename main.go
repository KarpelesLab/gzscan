package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	flagThreads = flag.Int("threads", runtime.NumCPU()*2, "change the number of threads used to scan the file")
)

func main() {
	flag.Parse()

	if err := doit(); err != nil {
		log.Printf("operation failed: %s", err)
		os.Exit(1)
	}
}

func doit() error {
	args := flag.Args()
	if len(args) != 1 {
		return errors.New("you need to provide a file to scan")
	}

	fil, err := os.Open(args[0])
	if err != nil {
		return fmt.Errorf("while opening %s: %w", args[0], err)
	}

	st, err := fil.Stat()
	if err != nil {
		return err
	}

	size := st.Size()
	threads := *flagThreads
	interval := size / int64(threads)
	var wg sync.WaitGroup
	errCh := make(chan error, threads)
	position := make([]uint64, threads)

	log.Printf("running with %d threads", threads)

	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			// run thread
			start := interval * int64(i)
			end := interval*int64(i) + interval + 256 // 256 bytes is enough overlap
			if end > size {
				// ensure we don't go too far
				end = size
			}

			err := performScan(i, position, fil, start, end)
			if err != nil {
				errCh <- err
			}
		}(i)
	}

	wg.Wait()

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

func performScan(thread int, position []uint64, fil *os.File, start, end int64) error {
	thR := &threadReader{
		atomicPos: &position[thread],
		pos:       start,
		end:       end,
		f:         fil,
	}
	r := bufio.NewReaderSize(thR, 2*1024*1024)
	pos := start
	var buf []byte
	var err error

	for {
		buf, err = r.Peek(1024*1024 + 1024)
		if err != nil && err != io.EOF {
			return err
		} else if err == io.EOF && len(buf) < 32 {
			// can't read anymore
			break
		}

		// locate beginning of gzip file
		p := bytes.Index(buf, []byte{0x1f, 0x8b, 0x08}) // 8=deflate (0~7 are reserved)
		if p == -1 {
			pos += 1024 * 1024
			r.Discard(1024 * 1024)
			continue
		}
		if p > 0 {
			pos += int64(p)
			r.Discard(p)
			continue
		}

		// buf starts with 0x1f 0x8b 0x08
		flags := buf[3]
		ts := binary.LittleEndian.Uint32(buf[4:8])
		compFlags := buf[8]
		os := buf[9]
		if (flags&0xe0 != 0) || (os > 13) || (compFlags & ^byte(6) != 0) {
			pos += 1
			r.Discard(1)
			continue
		}
		payload := buf[9:] // move buf forward
		var info []string

		info = append(info, fmt.Sprintf("os=%s", getOsName(os)))

		if flags&0x04 == 0x04 {
			xlen := binary.LittleEndian.Uint16(payload[:2])
			if xlen > 0 {
				xtra := payload[2 : xlen+2]
				info = append(info, fmt.Sprintf("extra=%s", hex.EncodeToString(xtra)))
			}
			payload = payload[xlen+2:]
		}
		if flags&0x08 == 0x08 {
			// we should have a null terminated filename here
			endpos := bytes.IndexByte(payload, 0)
			if endpos > 256 {
				// too long
				pos += 1
				r.Discard(1)
				continue
			}
			fn := payload[:endpos]
			payload = payload[endpos+1:]
			info = append(info, fmt.Sprintf("filename=%s", fn))
		}
		if flags&0x01 == 0x01 {
			info = append(info, "flag=FTEXT")
		}
		if flags&0x02 == 0x02 {
			info = append(info, "flag=FHCRC") // header checksum
		}
		if flags&0x10 == 0x10 {
			info = append(info, "flag=FCOMMENT")
		}

		log.Printf("found likely gzip: pos=%d stamp=%s %s", pos, time.Unix(int64(ts), 0), strings.Join(info, " "))

		pos += 3
		r.Discard(3)
	}
	return nil
}

func getOsName(os byte) string {
	switch os {
	case 0:
		return "FAT filesystem (MS-DOS, OS/2, NT/Win32)"
	case 1:
		return "Amiga"
	case 2:
		return "VMS (or OpenVMS)"
	case 3:
		return "Unix"
	case 4:
		return "VM/CMS"
	case 5:
		return "Atari TOS"
	case 6:
		return "HPFS filesystem (OS/2, NT)"
	case 7:
		return "Macintosh"
	case 8:
		return "Z-System"
	case 9:
		return "CP/M"
	case 10:
		return "TOPS-20"
	case 11:
		return "NTFS filesystem (NT)"
	case 12:
		return "QDOS"
	case 13:
		return "Acorn RISCOS"
	default:
		return "unknown"
	}
}
