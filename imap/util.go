package imap

import "github.com/emersion/go-imap"

func singleSeqSet(uid uint32) *imap.SeqSet {
	seq := new(imap.SeqSet)
	seq.AddNum(uid)
	return seq
}
