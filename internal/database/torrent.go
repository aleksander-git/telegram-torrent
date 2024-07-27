package database

import (
	"fmt"
	"strings"
)

const (
	InQueue     = "в очереди"
	Downloading = "скачивается"
	Loaded      = "загружен"
)

type Torrent struct {
	Link   string
	Status string
}

func (t Torrent) String() string {
	return fmt.Sprintf("`%s` - %s", t.Link, t.Status)
}

type TorrentList []Torrent

func (tl TorrentList) String() string {
	if len(tl) == 0 {
		return "У вас пока нет торрентов"
	}

	builder := strings.Builder{}
	builder.WriteString("Ваши торренты:\n")

	for i, v := range tl {
		builder.WriteString(fmt.Sprintf("\n%d) ", i+1))
		builder.WriteString(v.String())
	}

	return builder.String()
}
