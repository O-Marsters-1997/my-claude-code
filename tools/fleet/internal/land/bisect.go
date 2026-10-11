package land

import (
	"context"
	"fmt"

	"github.com/O-Marsters-1997/my-claude-code/tools/fleet/internal/deps"
)

type landedPR struct {
	number int
	sha    string
}

func bisect(ctx context.Context, d deps.Deps, in Input, landed []landedPR) (int, error) {
	lo, hi := 0, len(landed)-1
	for lo < hi {
		mid := (lo + hi) / 2
		fmt.Fprintf(d.Log, "bisect: #%d (%s)\n", landed[mid].number, landed[mid].sha)
		green, err := suiteGreen(ctx, d, in, landed[mid].sha)
		if err != nil {
			return 0, err
		}
		if green {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return landed[lo].number, nil
}
