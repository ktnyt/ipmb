package ipmb

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ipfs/go-ipfs-api"
	peer "github.com/libp2p/go-libp2p-peer"
)

type Relay struct {
	Ipfs   *shell.Shell
	ID     string
	Peers  map[peer.ID]time.Time
	Buffer map[peer.ID]string
	ping   *shell.PubSubSubscription
	data   *shell.PubSubSubscription
	level  uint
}

type RelayData struct {
	Time time.Time
	Data string
}

func NewRelay(ipfs *shell.Shell) (*Relay, error) {
	var err error

	var id *shell.IdOutput
	if id, err = ipfs.ID(); err != nil {
		return nil, fmt.Errorf("failed to initialize relay: %s", err)
	}

	return &Relay{
		Ipfs:   ipfs,
		ID:     id.ID,
		Peers:  make(map[peer.ID]time.Time),
		Buffer: make(map[peer.ID]string),
	}, nil
}

func (r *Relay) Join(level uint) (err error) {
	r.level = level

	ns := r.ID[2 : 2+r.level]

	pingTopic := fmt.Sprintf("/ipmb/relay/%s/ping", ns)
	dataTopic := fmt.Sprintf("/ipmb/relay/%s/data", ns)

	if r.ping, err = r.Ipfs.PubSubSubscribe(pingTopic); err != nil {
		return fmt.Errorf("relay failed to join '%s': %s", pingTopic, err)
	}

	if r.data, err = r.Ipfs.PubSubSubscribe(dataTopic); err != nil {
		return fmt.Errorf("relay failed to join '%s': %s", dataTopic, err)
	}

	return nil
}

func (r *Relay) Ping() (err error) {
	ns := r.ID[2 : 2+r.level]
	topic := fmt.Sprintf("/ipmb/relay/%s/ping", ns)

	var data []byte
	if data, err = time.Now().MarshalText(); err != nil {
		return fmt.Errorf("relay failed to ping '%s': %s", ns, err)
	}

	if err = r.Ipfs.PubSubPublish(topic, string(data)); err != nil {
		return fmt.Errorf("relay failed to ping '%s': %s", ns, err)
	}

	return nil
}

func (r *Relay) TestPing() (err error) {
	if r.ping == nil {
		return fmt.Errorf("relay failed to test next ping: not joined yet")
	}

	var rec shell.PubSubRecord
	if rec, err = r.ping.Next(); err != nil {
		return fmt.Errorf("relay failed to test next ping: %s", err)
	}

	data := rec.Data()

	if len(data) > 0 {
		var timestamp time.Time
		if timestamp.UnmarshalText(data) != nil {
			// If some garbage is being sent, discard it
			return nil
		}

		r.Peers[rec.From()] = timestamp
	}

	return nil
}

func (r *Relay) SendData(content string) (err error) {
	data := RelayData{
		Time: time.Now(),
		Data: content,
	}

	var jsonBytes []byte
	if jsonBytes, err = json.Marshal(data); err != nil {
		return fmt.Errorf("relay failed to send data: %s", err)
	}

	ns := r.ID[2 : 2+r.level]
	topic := fmt.Sprintf("/ipmb/relay/%s/data", ns)

	if err = r.Ipfs.PubSubPublish(topic, string(jsonBytes)); err != nil {
		return fmt.Errorf("relay failed to send data '%s': %s", ns, err)
	}

	return nil
}

func (r *Relay) TestData() (err error) {
	if r.data == nil {
		fmt.Errorf("relay failed to test next data: not joined yet")
	}

	var rec shell.PubSubRecord
	if rec, err = r.data.Next(); err != nil {
		return fmt.Errorf("relay failed to test next data: %s", err)
	}

	data := new(RelayData)

	if json.Unmarshal(rec.Data(), data) != nil {
		// Just forget about the message if garbage is sent
		return nil
	}

	from := rec.From()

	r.Peers[from] = data.Time
	r.Buffer[from] = data.Data

	return nil
}
