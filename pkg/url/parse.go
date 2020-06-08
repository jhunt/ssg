package url

import (
	"fmt"
	"regexp"
	"strings"
)

var re *regexp.Regexp

func init() {
	re = regexp.MustCompile(`^(.*?)://([^/].*?)(?:/(.*?)(/.*)?)?$`)
}

func Parse(s string) (*URL, error) {
	if m := re.FindStringSubmatch(s); m != nil {
		if strings.ToLower(m[1]) != "ssg" {
			return nil, fmt.Errorf("invalid scheme '%s'", m[1])
		}

		return &URL{
			Cluster: m[2],
			Bucket: m[3],
			Path: strings.Trim(m[4], "/"),
		}, nil
	}
	return nil, fmt.Errorf("invalid ssg url")
}

func (u URL) String() string {
	return "ssg://"+u.Cluster+"/"+u.Bucket+"/"+strings.Trim(u.Path, "/")
}
