package torrent

import (
	"fmt"
	"strings"
)

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
