package ipmb

import (
	"fmt"
	"strings"

	"github.com/ipfs/go-ipfs-api"
)

// Ipmb contains the information needed to operate on IPMB
type Ipmb struct {
	Ipfs      *shell.Shell
	Relay     *Relay
	ID        string
	Head      string
	Followees []string
}

func NewIpmb(url string) (*Ipmb, error) {
	var err error

	ipfs := shell.NewShell(url)

	var id *shell.IdOutput
	if id, err = ipfs.ID(); err != nil {
		return nil, fmt.Errorf("failed to initialize ipmb: %s", err)
	}

	var relay *Relay
	if relay, err = NewRelay(ipfs); err != nil {
		return nil, err
	}

	return &Ipmb{
		Ipfs:      ipfs,
		Relay:     relay,
		ID:        id.ID,
		Head:      "",
		Followees: []string{id.ID},
	}, nil
}

func (i *Ipmb) Post(status string) (err error) {
	var post string
	if post, err = i.Ipfs.NewObject("unixfs-dir"); err != nil {
		return fmt.Errorf("failed to create post: %s", err)
	}

	var content string
	if content, err = i.Ipfs.Add(strings.NewReader(status)); err != nil {
		return fmt.Errorf("failed to create post: %s", err)
	}

	if post, err = i.Ipfs.PatchLink(post, "content", content, false); err != nil {
		return fmt.Errorf("failed to create post: %s", err)
	}

	if len(i.Head) > 0 {
		if post, err = i.Ipfs.PatchLink(post, "previous", i.Head, false); err != nil {
			return fmt.Errorf("failed to create post: %s", err)
		}
	}

	i.Head = post

	return nil
}
