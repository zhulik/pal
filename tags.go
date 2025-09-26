package pal

import (
	"fmt"
	"strings"
)

type Tag string

const (
	TagSkip           Tag = "skip"
	TagMatchInterface Tag = "match_interface"
	TagName           Tag = "name"
)

var supportedTags = map[Tag]bool{
	TagSkip:           true,
	TagMatchInterface: true,
	TagName:           true,
}

func ParseTag(tags string) (map[Tag]string, error) {
	tagMap := make(map[Tag]string)
	tags = strings.ReplaceAll(tags, " ", "")
	if tags == "" {
		return tagMap, nil
	}

	for tag := range strings.SplitSeq(tags, ",") {
		parts := strings.Split(tag, "=")
		tagName := parts[0]

		if !supportedTags[Tag(tagName)] {
			return nil, fmt.Errorf("%w: tag unsupported %s", ErrInvalidTag, tagName)
		}
		switch len(parts) {
		case 2:
			if parts[1] == "" {
				return nil, fmt.Errorf("%w: tag is malformed %s", ErrInvalidTag, tag)
			}
			tagMap[Tag(tagName)] = parts[1]
		case 1:
			tagMap[Tag(tagName)] = ""
		default:
			return nil, fmt.Errorf("%w: tag is malformed %s", ErrInvalidTag, tag)
		}
	}
	return tagMap, nil
}
