package ipmb

import (
	"fmt"
	"strings"

	"github.com/ipfs/go-ipfs-api"
)

type Ipmb struct {
	Ipfs      *shell.Shell
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

	return &Ipmb{
		Ipfs:      ipfs,
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

func (i *Ipmb) Notify() (err error) {
	topic := fmt.Sprintf("/ipmb/%s/head", i.ID)
	fmt.Println(topic)
	return i.Ipfs.PubSubPublish(topic, i.Head)
}

func (i *Ipmb) Watch(c chan shell.PubSubRecord, q chan bool) (err error) {
	topic := fmt.Sprintf("/ipmb/%s/head", i.ID)

	var sub *shell.PubSubSubscription
	if sub, err = i.Ipfs.PubSubSubscribe(topic); err != nil {
		return err
	}

	for {
		var rec shell.PubSubRecord
		if rec, err = sub.Next(); err != nil {
			return err
		}

		if len(string(rec.Data())) > 0 {
			select {
			case c <- rec:
				continue
			case <-q:
				return nil
			}
		}
	}
}
