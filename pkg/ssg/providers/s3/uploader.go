package s3

import (
	"github.com/jhunt/go-s3"
)

type Uploader struct {
	key string
	up  *s3.Upload
	n   int64

	bufn int
	buf  []byte
}

func (out *Uploader) Write(b []byte) (int, error) {
	// calculate the amount of space left in our send buffer.
	left := len(out.buf) - out.bufn

	nwrit := 0
	for len(b) >= left {
		// fill up our send buffer, so that we get a complete
		// multi-part of the correct segment size.
		copy(out.buf[out.bufn:], b[:left])

		// write our full multi-part to the backend s3 store.
		if err := out.up.Write(out.buf); err != nil {
			return nwrit, err
		}

		// track the new data we wrote directly.
		nwrit += left

		// slide our input buffer back to account for the
		// direct write.
		b = b[left:]

		// our send buffer is now empty, ready to be re-filled.
		left = len(out.buf)
		out.bufn = 0
	}

	// place the leftover input data into our send buffer
	// for a future call to Write() or Close().
	copy(out.buf[out.bufn:], b)
	out.bufn += len(b)

	// record the send-buffered remainder of the input buffer
	// as having been written (return its byte counts) since
	// the send buffer cache is "invisible" to callers.
	nwrit += len(b)
	out.n += int64(nwrit)
	return nwrit, nil
}

func (out *Uploader) Close() error {
	if out.bufn > 0 {
		if err := out.up.Write(out.buf[:out.bufn]); err != nil {
			return err
		}
		out.n += int64(out.bufn)
	}
	return out.up.Done()
}

func (out *Uploader) WroteCompressed() int64 {
	return out.n
}

func (out *Uploader) WroteUncompressed() int64 {
	return out.n
}

func (out *Uploader) Path() string {
	return out.key
}

func (out *Uploader) Cancel() error {
	return nil
}
