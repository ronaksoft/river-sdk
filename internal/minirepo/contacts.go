package minirepo

/*
   Creation Time: 2021 - May - 05
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

const (
	prefixContacts = "CONTACTS"
	indexContacts
)

var (
	bucketContacts = []byte("DLG")
)

type repoContacts struct {
	*repository
}

func newContact(r *repository) *repoContacts {
	rd := &repoContacts{
		repository: r,
	}
	return rd
}
