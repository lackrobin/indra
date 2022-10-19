package dialog

import (
	"sync"

	"github.com/Indra-Labs/indra/pkg/fing"
	"github.com/Indra-Labs/indra/pkg/keys"
	"github.com/Indra-Labs/indra/pkg/mesg"
	"github.com/Indra-Labs/indra/pkg/sifr"
)

// Dialog is a data structure for tracking keys used in a message exchange.
type Dialog struct {
	sync.Mutex
	// LastIn is the newest pubkey seen in a received message from the
	// correspondent.
	LastIn *keys.Pubkey
	// LastOut is the newest privkey used in an outbound message.
	LastOut *keys.Privkey
	// Seen are the keys that have been seen since the last new message sent
	// out to the correspondent.
	Seen []*keys.Pubkey
	// Used are the recently used keys that have not been invalidated by the
	// counterparty sending them in the Expires field.
	Used []*keys.Privkey
	// UsedFingerprints are 1:1 mapped to Used private keys for fast
	// recognition. These have been sent in Expires field.
	UsedFingerprints []fing.Fingerprint
	// SegmentSize is the size of packets used in the Dialog. Anything
	// larger will be segmented and potentially augmented with Reed Solomon
	// parity shards for retransmit avoidance.
	SegmentSize uint16
}

// New creates a new Dialog for tracking a conversation between two nodes.
// For the initiator, the pubkey is the current one advertised by the
// correspondent, and for a correspondent, this pubkey is from the first one
// appearing in the initial message.
func New(pub *keys.Pubkey) (d *Dialog) {
	d = &Dialog{LastIn: pub}
	return
}

// Frame is the data format that goes on the wire. This message is wrapped
// inside a Message and the payload is also inside a Message.
type Frame struct {
	// To is the fingerprint of the pubkey used in the ECDH key exchange.
	To *fing.Fingerprint
	// From is the pubkey corresponding to the private key used in the ECDH
	// key exchange.
	From *keys.PubkeyBytes
	// Expires are the fingerprints of public keys that the correspondent
	// can now discard as they will not be used again.
	Expires []fing.Fingerprint
	// Seen are all the keys excluding the To key to signal these can be
	// deleted.
	Seen []fing.Fingerprint
	// Seq specifies the segment number of the message.
	Seq uint32
	// Data is a Crypt containing a Message.
	Data *sifr.Crypt
}

// Send issues a new message.
func (d *Dialog) Send(payload []byte) (wf *Frame, e error) {
	// generate the sender private key
	var prv *keys.Privkey
	if prv, e = keys.GeneratePrivkey(); log.I.Chk(e) {
		return
	}
	pub := prv.Pubkey()
	wf = &Frame{}
	// Fill in the 'From' key to the pubkey of the new private key.
	wf.From = pub.Serialize()
	// Lock the mutex of Dialog, so we can update the used/seen keys.
	d.Mutex.Lock()
	// We always send new messages to the last known correspondent pubkey.
	lastin := d.LastIn
	// Move the last outbound private key into the Used field.
	if d.LastOut != nil {
		d.Used = append(d.Used, d.LastOut)
	}
	// Set current key to the last used.
	d.LastOut = prv
	// Collect the used keys to put in the expired. These will be deleted
	// in the receiver function.
	if len(d.Used) > 0 {
		for i := range d.Used {
			fp := d.Used[i].Pubkey().Fingerprint()
			wf.Expires = append(wf.Expires, fp)
		}
	}
	// Seen keys signal to the correspondent they can discard the related
	// private key as it will not be addressed to again.
	if len(d.Seen) > 0 {
		for i := range d.Seen {
			wf.Seen = append(wf.Seen, d.Seen[i].Fingerprint())
		}
	}
	// This is the last access on the Dialog, so we can unlock here.
	d.Mutex.Unlock()
	// Getting secret and To here outside the critical section as it
	// doesn't need locking once the pubkey is copied.
	secret := prv.ECDH(lastin)
	tofp := lastin.Fingerprint()
	wf.To = &tofp
	var msg *mesg.Message
	if msg, e = mesg.New(payload, prv); log.E.Chk(e) {
		return
	}
	var em *sifr.Crypt
	if em, e = sifr.NewCrypt(msg, secret); log.E.Chk(e) {
		return
	}
	wf.Data = em
	return
}

// Receive processes a received message, handles expiring correspondent and
// prior send keys, and returns the decrypted message to the caller.
func (d *Dialog) Receive(message []byte) (m *mesg.Message, e error) {
	// Lock the mutex of Dialog, so we can update the used/seen keys.
	d.Mutex.Lock()
	// This is the last access on the Dialog, so we can unlock here.
	d.Mutex.Unlock()
	return
}
