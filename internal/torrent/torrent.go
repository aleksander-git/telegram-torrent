package torrent

import (
	"fmt"

	"github.com/go-bittorrent/magneturi"
)

const (
	InQueue     = "в очереди"
	Downloading = "скачивается"
	Loaded      = "загружен"
)

type TorrentParseError struct {
	message string
	link    string
}

func (tpe TorrentParseError) Error() string {
	return fmt.Sprintf("torrent link %q is invalid: %s", tpe.link, tpe.message)
}

type Torrent struct {
	Link   string
	Status string
}

func New(link string) (Torrent, error) {
	if _, err := magneturi.Parse(link); err != nil {
		return Torrent{}, TorrentParseError{
			message: err.Error(),
			link:    link,
		}
	}

	return Torrent{
		Link:   link,
		Status: InQueue,
	}, nil
}

func (t Torrent) String() string {
	return fmt.Sprintf("`%s` - %s", t.Link, t.Status)
}
