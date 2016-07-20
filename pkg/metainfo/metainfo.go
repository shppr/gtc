// metainfo a package for dealing with '.torrent' files
package metainfo

import (
	"crypto/sha1"
	"fmt"
	"os"
	"time"

	"github.com/marksamman/bencode"
)

// MetaInfo a mapping of a .torrent file to a struct
type MetaInfo struct {
	Info
	Announce     string
	AnnounceList [][]string
	CreationDate time.Time
	Comment      string
	CreatedBy    string
	Encoding     string
	Files        []File
	Name         string // Single File - name, Multi file - dirname
	InfoHash     [20]byte
}

// Info fields common to both single and multi file info dictionary
type Info struct {
	PieceLength int64
	Pieces      []byte
	Private     bool
}

// File a struct for files in multifileinfo dicts
type File struct {
	Length int64
	MD5Sum []byte
	Path   []string
}

func NewFromFilename(fn string) (*MetaInfo, error) {
	file, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := bencode.Decode(file)
	if err != nil {
		return nil, err
	}

	m := &MetaInfo{}

	// Populate Announce or AnnounceList
	annLists, ok := data["announce-list"].([]interface{})
	lists := [][]string{}
	if !ok {
		m.Announce = data["announce"].(string)
	}
	for _, list := range annLists {
		al := []string{}
		for _, URL := range list.([]interface{}) {
			al = append(al, URL.(string))
		}
		lists = append(lists, al)
	}
	m.AnnounceList = lists

	// parse additional optional fields
	if cd, ok := data["creation date"]; ok {
		m.CreationDate = time.Unix(cd.(int64), 0)
	}
	if c, ok := data["comment"]; ok {
		m.Comment = c.(string)
	}
	if cb, ok := data["created by"]; ok {
		m.CreatedBy = cb.(string)
	}
	if enc, ok := data["encoding"]; ok {
		m.Encoding = enc.(string)
	}

	// begin populating the Info dict
	info := data["info"].(map[string]interface{})
	infoHash := sha1.Sum(bencode.Encode(info))
	m.InfoHash = infoHash

	if name, ok := info["name"]; ok {
		m.Name = name.(string)
	}

	if plen, ok := info["piece length"]; ok {
		m.PieceLength = plen.(int64)
	}

	if pieces, ok := info["pieces"]; ok {
		m.Pieces = []byte(pieces.(string))
	}

	if private, ok := info["private"]; ok {
		if private.(int64) == 1 {
			m.Private = true
		} else {
			m.Private = false
		}
	}

	if files, exists := info["files"]; !exists {
		//TODO name and length arent optional, maybe return an err here
		f := File{}

		if length, ok := info["length"]; ok {
			f.Length = length.(int64)
		}
		if md5, ok := info["md5sum"]; ok {
			f.MD5Sum = md5.([]byte)
		}

		m.Files = append(m.Files, f)
	} else {
		for _, file := range files.([]interface{}) {
			f := File{}
			fileDict := file.(map[string]interface{})
			if length, ok := fileDict["length"]; ok {
				f.Length = length.(int64)
			}

			if path, ok := fileDict["path"]; ok {
				var pathTokens []string
				for _, p := range path.([]interface{}) {
					pathTokens = append(pathTokens, p.(string))
				}
				f.Path = pathTokens
			}

			if md5, ok := fileDict["md5sum"]; ok {
				f.MD5Sum = md5.([]byte)
			}

			m.Files = append(m.Files, f)
		}
	}

	return m, nil
}

func (m *MetaInfo) String() string {
	ret := fmt.Sprintf("Announce: %v\n", m.Announce)
	ret += fmt.Sprintf("AnnounceList(opt): %v\n", m.AnnounceList)
	ret += fmt.Sprintf("Creation Date(opt): %v\n", m.CreationDate)
	ret += fmt.Sprintf("Comment(opt): %v\n", m.Comment)
	ret += fmt.Sprintf("Created By(opt): %v\n", m.CreatedBy)
	ret += fmt.Sprintf("Encoding(opt): %v\n", m.Encoding)
	ret += fmt.Sprintf("Name: %v\n", m.Name)
	ret += fmt.Sprintf("Info: \n")
	if l := len(m.Files); l == 1 {
		ret += fmt.Sprintf("    Length: %v\n", m.Files[0].Length)
		ret += fmt.Sprintf("    MD5(opt): %v\n", m.Files[0].MD5Sum)
	} else if l > 1 {
		for _, f := range m.Files {
			ret += fmt.Sprintf("    Filename: %v\n", f.Path[len(f.Path)-1:])
			ret += fmt.Sprintf("    Length: %v\n", f.Length)
			ret += fmt.Sprintf("    MD5(opt): %v\n", f.MD5Sum)
		}
	}
	ret += fmt.Sprintf("Pieces: %v\n", len(m.Pieces))
	ret += fmt.Sprintf("Piece Len: %v\n", m.PieceLength)
	ret += fmt.Sprintf("Private: %v\n", m.Private)
	return ret
}
