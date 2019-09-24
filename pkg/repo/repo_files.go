package repo

import (
	"fmt"
	msg "git.ronaksoftware.com/ronak/riversdk/msg/ext"
	ronak "git.ronaksoftware.com/ronak/toolbox"
	"github.com/dgraph-io/badger"
	"time"
)

/*
   Creation Time: 2019 - Sep - 15
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/

const (
	prefixFiles = "FILES"
)

type repoFiles struct {
	*repository
}

func (r *repoFiles) getKey(clusterID int32, fileID int64, accessHash uint64) []byte {
	return ronak.StrToByte(fmt.Sprintf("%s.%012d.%021d.%021d", prefixFiles, clusterID, fileID, accessHash))
}

func (r *repoFiles) Get(clusterID int32, fileID int64, accessHash uint64) (*msg.ClientFile, error) {
	file := new(msg.ClientFile)
	err := r.badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get(r.getKey(clusterID, fileID, accessHash))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return file.Unmarshal(val)
		})
	})
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (r *repoFiles) GetMediaDocument(m *msg.UserMessage) (*msg.ClientFile, error) {
	md := new(msg.MediaDocument)
	_ = md.Unmarshal(m.Media)
	return r.Get(md.Doc.ClusterID, md.Doc.ID, md.Doc.AccessHash)
}

func (r *repoFiles) SaveUserPhotos(u *msg.User) error {
	photos := make([]*msg.UserPhoto, 1+len(u.PhotoGallery))
	if u.Photo != nil {
		photos = append(photos, u.Photo)
	}
	for _, photo := range u.PhotoGallery {
		photos = append(photos, photo)
	}
	for _, photo := range photos {
		_ = ronak.Try(100, time.Millisecond, func() error {
			return r.Save(&msg.ClientFile{
				ClusterID:  photo.PhotoBig.ClusterID,
				FileID:     photo.PhotoBig.FileID,
				AccessHash: photo.PhotoBig.AccessHash,
				Type:       msg.ClientFileType_AccountProfilePhoto,
				MimeType:   "",
				UserID:     u.ID,
				GroupID:    0,
				FileSize:   0,
				MessageID:  0,
				PeerID:     u.ID,
				PeerType:   int32(msg.PeerUser),
				Version:    0,
			})
		})
		_ = ronak.Try(100, time.Millisecond, func() error {
			return r.Save(&msg.ClientFile{
				ClusterID:  photo.PhotoSmall.ClusterID,
				FileID:     photo.PhotoSmall.FileID,
				AccessHash: photo.PhotoSmall.AccessHash,
				Type:       msg.ClientFileType_Thumbnail,
				MimeType:   "",
				UserID:     u.ID,
				GroupID:    0,
				FileSize:   0,
				MessageID:  0,
				PeerID:     u.ID,
				PeerType:   int32(msg.PeerUser),
				Version:    0,
			})
		})

	}
	return nil
}

func (r *repoFiles) SaveContactPhoto(u *msg.ContactUser) error {
	if u.Photo == nil {
		return nil
	}
	err := r.Save(&msg.ClientFile{
		ClusterID:  u.Photo.PhotoBig.ClusterID,
		FileID:     u.Photo.PhotoBig.FileID,
		AccessHash: u.Photo.PhotoBig.AccessHash,
		Type:       msg.ClientFileType_AccountProfilePhoto,
		MimeType:   "",
		UserID:     u.ID,
		GroupID:    0,
		FileSize:   0,
		MessageID:  0,
		PeerID:     u.ID,
		PeerType:   int32(msg.PeerUser),
		Version:    0,
	})
	if err != nil {
		return err
	}
	err = r.Save(&msg.ClientFile{
		ClusterID:  u.Photo.PhotoSmall.ClusterID,
		FileID:     u.Photo.PhotoSmall.FileID,
		AccessHash: u.Photo.PhotoSmall.AccessHash,
		Type:       msg.ClientFileType_Thumbnail,
		MimeType:   "",
		UserID:     u.ID,
		GroupID:    0,
		FileSize:   0,
		MessageID:  0,
		PeerID:     u.ID,
		PeerType:   int32(msg.PeerUser),
		Version:    0,
	})
	return err
}

func (r *repoFiles) SaveGroupPhoto(g *msg.Group) error {
	if g.Photo == nil {
		return nil
	}
	err := r.Save(&msg.ClientFile{
		ClusterID:  g.Photo.PhotoBig.ClusterID,
		FileID:     g.Photo.PhotoBig.FileID,
		AccessHash: g.Photo.PhotoBig.AccessHash,
		Type:       msg.ClientFileType_GroupProfilePhoto,
		MimeType:   "",
		UserID:     0,
		GroupID:    g.ID,
		FileSize:   0,
		MessageID:  0,
		PeerID:     g.ID,
		PeerType:   int32(msg.PeerGroup),
		Version:    0,
	})
	if err != nil {
		return err
	}
	err = r.Save(&msg.ClientFile{
		ClusterID:  g.Photo.PhotoSmall.ClusterID,
		FileID:     g.Photo.PhotoSmall.FileID,
		AccessHash: g.Photo.PhotoSmall.AccessHash,
		Type:       msg.ClientFileType_Thumbnail,
		MimeType:   "",
		UserID:     0,
		GroupID:    g.ID,
		FileSize:   0,
		MessageID:  0,
		PeerID:     g.ID,
		PeerType:   int32(msg.PeerGroup),
		Version:    0,
	})
	return err
}

func (r *repoFiles) SaveMessageMedia(m *msg.UserMessage) error {
	switch m.MediaType {
	case msg.MediaTypeDocument:
		md := new(msg.MediaDocument)
		err := md.Unmarshal(m.Media)
		if err != nil {
			return err
		}

		err = r.Save(&msg.ClientFile{
			ClusterID:  md.Doc.ClusterID,
			FileID:     md.Doc.ID,
			AccessHash: md.Doc.AccessHash,
			Type:       msg.ClientFileType_Message,
			MimeType:   md.Doc.MimeType,
			UserID:     0,
			GroupID:    0,
			FileSize:   int64(md.Doc.FileSize),
			MessageID:  m.ID,
			PeerID:     m.PeerID,
			PeerType:   m.PeerType,
			Version:    md.Doc.Version,
		})
		if err != nil {
			return err
		}

		if md.Doc.Thumbnail != nil {
			err = r.Save(&msg.ClientFile{
				ClusterID:  md.Doc.Thumbnail.ClusterID,
				FileID:     md.Doc.Thumbnail.FileID,
				AccessHash: md.Doc.Thumbnail.AccessHash,
				Type:       msg.ClientFileType_Thumbnail,
				MimeType:   "",
				UserID:     0,
				GroupID:    0,
				FileSize:   0,
				MessageID:  m.ID,
				PeerID:     m.PeerID,
				PeerType:   m.PeerType,
				Version:    0,
			})
			return err
		}
	}

	return nil
}

func (r *repoFiles) Save(file *msg.ClientFile) error {
	if file == nil {
		return nil
	}

	fileBytes, _ := file.Marshal()

	return r.badger.Update(func(txn *badger.Txn) error {
		err := txn.SetEntry(badger.NewEntry(
			r.getKey(file.ClusterID, file.FileID, file.AccessHash),
			fileBytes,
		))
		if err != nil {
			return err
		}
		return nil
	})
}

func (r *repoFiles) Delete(clusterID int32, fileID int64, accessHash uint64) error {
	return r.badger.Update(func(txn *badger.Txn) error {
		return txn.Delete(r.getKey(clusterID, fileID, accessHash))
	})
}
