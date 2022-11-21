package post

import (
	"context"
	"net/http"
	"time"

	"github.com/uhppoted/uhppoted-httpd/system"
)

func SynchronizeACL(ctx context.Context, w http.ResponseWriter, r *http.Request, timeout time.Duration) {
	ch := make(chan struct{})
	ctx, cancel := context.WithTimeout(ctx, timeout)

	defer cancel()

	go func() {
		debugf("ACL", "synchronizing ACL")
		if err := system.SynchronizeACL(); err != nil {
			warnf("ACL", "%v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			infof("ACL", "ACL synchronized")
		}

		close(ch)
	}()

	select {
	case <-ctx.Done():
		warnf("ACL", "%v", ctx.Err())
		http.Error(w, "Timeout waiting for response from system", http.StatusInternalServerError)

	case <-ch:
	}
}
