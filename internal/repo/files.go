package repo

import (
    "context"
    "fmt"
    "mime"
    "os"
    "path"
    "path/filepath"
    "strings"
    "sync"

    "github.com/dgraph-io/badger/v2"
    "github.com/dgraph-io/badger/v2/pb"
    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/domain"
    "github.com/ronaksoft/rony/tools"
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
    prefixFiles         = "FILES"
    prefixFilesRequests = "FILES_REQ"
)

type repoFiles struct {
    *repository
}

func getFileKey(clusterID int32, fileID int64, accessHash uint64) []byte {
    return tools.StrToByte(fmt.Sprintf("%s.%012d.%021d.%021d", prefixFiles, clusterID, fileID, accessHash))
}

func getFile(txn *badger.Txn, clusterID int32, fileID int64, accessHash uint64) (*msg.ClientFile, error) {
    file := &msg.ClientFile{}
    item, err := txn.Get(getFileKey(clusterID, fileID, accessHash))
    if err != nil {
        return nil, err
    }
    err = item.Value(func(val []byte) error {
        return file.Unmarshal(val)
    })
    if err != nil {
        return nil, err
    }
    return file, nil
}

func saveFile(txn *badger.Txn, file *msg.ClientFile) error {
    fileBytes, _ := file.Marshal()
    return txn.SetEntry(badger.NewEntry(
        getFileKey(file.ClusterID, file.FileID, file.AccessHash),
        fileBytes,
    ))
}

func saveUserPhotos(txn *badger.Txn, userID int64, photos ...*msg.UserPhoto) error {
    for _, photo := range photos {
        if photo == nil {
            continue
        }
        if photo.PhotoBig != nil {
            err := saveFile(txn, &msg.ClientFile{
                ClusterID:  photo.PhotoBig.ClusterID,
                FileID:     photo.PhotoBig.FileID,
                AccessHash: photo.PhotoBig.AccessHash,
                Type:       msg.ClientFileType_AccountProfilePhoto,
                MimeType:   "",
                UserID:     userID,
                GroupID:    0,
                FileSize:   0,
                MessageID:  0,
                PeerID:     userID,
                PeerType:   int32(msg.PeerType_PeerUser),
                Version:    0,
            })
            if err != nil {
                return err
            }
        }
        if photo.PhotoSmall != nil {
            err := saveFile(txn, &msg.ClientFile{
                ClusterID:  photo.PhotoSmall.ClusterID,
                FileID:     photo.PhotoSmall.FileID,
                AccessHash: photo.PhotoSmall.AccessHash,
                Type:       msg.ClientFileType_Thumbnail,
                MimeType:   "",
                UserID:     userID,
                GroupID:    0,
                FileSize:   0,
                MessageID:  0,
                PeerID:     userID,
                PeerType:   int32(msg.PeerType_PeerUser),
                Version:    0,
            })
            if err != nil {
                return err
            }
        }

        bytes, _ := photo.Marshal()
        err := txn.SetEntry(badger.NewEntry(getUserPhotoGalleryKey(userID, photo.PhotoID), bytes))
        if err != nil {
            return err
        }
    }
    return nil
}

func deleteAllUserPhotos(txn *badger.Txn, userID int64) error {
    opts := badger.DefaultIteratorOptions
    opts.Prefix = getUserPhotoGalleryPrefix(userID)
    it := txn.NewIterator(opts)
    for it.Rewind(); it.Valid(); it.Next() {
        _ = txn.Delete(it.Item().KeyCopy(nil))
    }
    it.Close()
    return nil
}

func saveGroupPhotos(txn *badger.Txn, groupID int64, photos ...*msg.GroupPhoto) error {
    for _, photo := range photos {
        if photo != nil {
            if photo.PhotoBig != nil {
                err := saveFile(txn, &msg.ClientFile{
                    ClusterID:  photo.PhotoBig.ClusterID,
                    FileID:     photo.PhotoBig.FileID,
                    AccessHash: photo.PhotoBig.AccessHash,
                    Type:       msg.ClientFileType_GroupProfilePhoto,
                    MimeType:   "",
                    UserID:     0,
                    GroupID:    groupID,
                    FileSize:   0,
                    MessageID:  0,
                    PeerID:     groupID,
                    PeerType:   int32(msg.PeerType_PeerGroup),
                    Version:    0,
                })
                if err != nil {
                    return err
                }
            }
            if photo.PhotoSmall != nil {
                err := saveFile(txn, &msg.ClientFile{
                    ClusterID:  photo.PhotoSmall.ClusterID,
                    FileID:     photo.PhotoSmall.FileID,
                    AccessHash: photo.PhotoSmall.AccessHash,
                    Type:       msg.ClientFileType_Thumbnail,
                    MimeType:   "",
                    UserID:     0,
                    GroupID:    groupID,
                    FileSize:   0,
                    MessageID:  0,
                    PeerID:     groupID,
                    PeerType:   int32(msg.PeerType_PeerGroup),
                    Version:    0,
                })
                if err != nil {
                    return err
                }
            }
            bytes, _ := photo.Marshal()
            err := txn.SetEntry(badger.NewEntry(getGroupPhotoGalleryKey(groupID, photo.PhotoID), bytes))
            if err != nil {
                return err
            }
        }
    }
    return nil
}

func saveMessageMedia(txn *badger.Txn, m *msg.UserMessage) error {
    switch m.MediaType {
    case msg.MediaType_MediaTypeDocument:
        md := new(msg.MediaDocument)
        err := md.Unmarshal(m.Media)
        if err != nil {
            return err
        }

        fileExt := ""
        for _, attr := range md.Doc.Attributes {
            if attr.Type == msg.DocumentAttributeType_AttributeTypeFile {
                x := &msg.DocumentAttributeFile{}
                _ = x.Unmarshal(attr.Data)
                fileExt = filepath.Ext(x.Filename)
            }
        }

        err = saveFile(txn, &msg.ClientFile{
            ClusterID:   md.Doc.ClusterID,
            FileID:      md.Doc.ID,
            AccessHash:  md.Doc.AccessHash,
            Type:        msg.ClientFileType_Message,
            MimeType:    md.Doc.MimeType,
            Extension:   fileExt,
            UserID:      0,
            GroupID:     0,
            FileSize:    int64(md.Doc.FileSize),
            MessageID:   m.ID,
            PeerID:      m.PeerID,
            PeerType:    m.PeerType,
            Version:     md.Doc.Version,
            MD5Checksum: md.Doc.MD5Checksum,
            Attributes:  md.Doc.Attributes,
        })
        if err != nil {
            return err
        }

        if md.Doc.Thumbnail != nil {
            err = saveFile(txn, &msg.ClientFile{
                ClusterID:  md.Doc.Thumbnail.ClusterID,
                FileID:     md.Doc.Thumbnail.FileID,
                AccessHash: md.Doc.Thumbnail.AccessHash,
                Type:       msg.ClientFileType_Thumbnail,
                MimeType:   "jpeg",
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

func (r *repoFiles) SaveMessageMediaDocument(md *msg.MediaDocument) error {
    return badgerUpdate(func(txn *badger.Txn) error {
        return r.saveMessageMediaDocument(txn, md)
    })
}

func (r *repoFiles) saveMessageMediaDocument(txn *badger.Txn, md *msg.MediaDocument) error {
    fileExt := ""
    for _, attr := range md.Doc.Attributes {
        if attr.Type == msg.DocumentAttributeType_AttributeTypeFile {
            x := &msg.DocumentAttributeFile{}
            _ = x.Unmarshal(attr.Data)
            fileExt = filepath.Ext(x.Filename)
        }
    }

    err := saveFile(txn, &msg.ClientFile{
        ClusterID:   md.Doc.ClusterID,
        FileID:      md.Doc.ID,
        AccessHash:  md.Doc.AccessHash,
        Type:        msg.ClientFileType_Message,
        MimeType:    md.Doc.MimeType,
        Extension:   fileExt,
        UserID:      0,
        GroupID:     0,
        FileSize:    int64(md.Doc.FileSize),
        Version:     md.Doc.Version,
        MD5Checksum: md.Doc.MD5Checksum,
        Attributes:  md.Doc.Attributes,
    })
    if err != nil {
        return err
    }

    if md.Doc.Thumbnail != nil {
        err = saveFile(txn, &msg.ClientFile{
            ClusterID:  md.Doc.Thumbnail.ClusterID,
            FileID:     md.Doc.Thumbnail.FileID,
            AccessHash: md.Doc.Thumbnail.AccessHash,
            Type:       msg.ClientFileType_Thumbnail,
            MimeType:   "jpeg",
            UserID:     0,
            GroupID:    0,
            FileSize:   0,
            Version:    0,
        })
        return err
    }
    return nil
}

func (r *repoFiles) SaveWallpaper(txn *badger.Txn, wallpaper *msg.WallPaper) error {
    if wallpaper.Document == nil {
        return nil
    }

    fileExt := ""
    for _, attr := range wallpaper.Document.Attributes {
        if attr.Type == msg.DocumentAttributeType_AttributeTypeFile {
            x := &msg.DocumentAttributeFile{}
            _ = x.Unmarshal(attr.Data)
            fileExt = filepath.Ext(x.Filename)
        }
    }

    err := saveFile(txn, &msg.ClientFile{
        ClusterID:   wallpaper.Document.ClusterID,
        FileID:      wallpaper.Document.ID,
        AccessHash:  wallpaper.Document.AccessHash,
        Type:        msg.ClientFileType_Wallpaper,
        MimeType:    wallpaper.Document.MimeType,
        Extension:   fileExt,
        UserID:      0,
        GroupID:     0,
        FileSize:    int64(wallpaper.Document.FileSize),
        WallpaperID: wallpaper.ID,
        Version:     wallpaper.Document.Version,
        MD5Checksum: wallpaper.Document.MD5Checksum,
    })

    if err != nil {
        return err
    }

    if wallpaper.Document.Thumbnail != nil {
        err = saveFile(txn, &msg.ClientFile{
            ClusterID:   wallpaper.Document.Thumbnail.ClusterID,
            FileID:      wallpaper.Document.Thumbnail.FileID,
            AccessHash:  wallpaper.Document.Thumbnail.AccessHash,
            Type:        msg.ClientFileType_Thumbnail,
            MimeType:    "jpeg",
            UserID:      0,
            GroupID:     0,
            FileSize:    0,
            WallpaperID: wallpaper.ID,
            Version:     0,
        })
        return err
    }

    return nil
}

func (r *repoFiles) SaveGif(mediaDocument *msg.MediaDocument) error {
    if mediaDocument == nil || mediaDocument.Doc == nil {
        return nil
    }

    fileExt := ""
    for _, attr := range mediaDocument.Doc.Attributes {
        if attr.Type == msg.DocumentAttributeType_AttributeTypeFile {
            x := &msg.DocumentAttributeFile{}
            _ = x.Unmarshal(attr.Data)
            fileExt = filepath.Ext(x.Filename)
        }
    }

    err := badgerUpdate(func(txn *badger.Txn) error {
        err := saveFile(txn, &msg.ClientFile{
            ClusterID:   mediaDocument.Doc.ClusterID,
            FileID:      mediaDocument.Doc.ID,
            AccessHash:  mediaDocument.Doc.AccessHash,
            Type:        msg.ClientFileType_Gif,
            MimeType:    mediaDocument.Doc.MimeType,
            Extension:   fileExt,
            UserID:      0,
            GroupID:     0,
            FileSize:    int64(mediaDocument.Doc.FileSize),
            WallpaperID: 0,
            Version:     mediaDocument.Doc.Version,
            MD5Checksum: mediaDocument.Doc.MD5Checksum,
        })

        if err != nil {
            return err
        }

        if mediaDocument.Doc.Thumbnail != nil {
            err = saveFile(txn, &msg.ClientFile{
                ClusterID:   mediaDocument.Doc.Thumbnail.ClusterID,
                FileID:      mediaDocument.Doc.Thumbnail.FileID,
                AccessHash:  mediaDocument.Doc.Thumbnail.AccessHash,
                Type:        msg.ClientFileType_Thumbnail,
                MimeType:    "jpeg",
                UserID:      0,
                GroupID:     0,
                FileSize:    0,
                WallpaperID: 0,
                Version:     0,
            })
            return err
        }

        return nil
    })

    return err
}

func (r *repoFiles) Get(clusterID int32, fileID int64, accessHash uint64) (file *msg.ClientFile, err error) {
    err = badgerView(func(txn *badger.Txn) error {
        file, err = getFile(txn, clusterID, fileID, accessHash)
        return err
    })
    return
}

func (r *repoFiles) GetMediaDocument(m *msg.UserMessage) (*msg.ClientFile, error) {
    md := &msg.MediaDocument{}
    _ = md.Unmarshal(m.Media)
    return r.Get(md.Doc.ClusterID, md.Doc.ID, md.Doc.AccessHash)
}

func (r *repoFiles) Save(file *msg.ClientFile) error {
    if file == nil {
        return nil
    }
    return badgerUpdate(func(txn *badger.Txn) error {
        return saveFile(txn, file)
    })
}

func (r *repoFiles) Delete(clusterID int32, fileID int64, accessHash uint64) error {
    return badgerUpdate(func(txn *badger.Txn) error {
        return txn.Delete(getFileKey(clusterID, fileID, accessHash))
    })
}

func (r *repoFiles) GetCachedMedia(teamID int64) *msg.ClientCachedMediaInfo {
    userMediaInfo := make(map[int64]map[msg.ClientMediaType]int64, 128)
    groupMediaInfo := make(map[int64]map[msg.ClientMediaType]int64, 128)
    userMtx := sync.Mutex{}
    groupMtx := sync.Mutex{}

    stream := r.badger.NewStream()
    stream.Prefix = tools.StrToByte(fmt.Sprintf("%s.", prefixMessages))
    stream.ChooseKey = func(item *badger.Item) bool {
        m := &msg.UserMessage{}
        err := item.Value(func(val []byte) error {
            return m.Unmarshal(val)
        })
        if err != nil {
            return false
        }
        if m.TeamID != teamID {
            return false
        }
        switch m.MediaType {
        case msg.MediaType_MediaTypeDocument:
            d := msg.MediaDocument{}
            err = d.Unmarshal(m.Media)
            if err != nil {
                return false
            }

            f, err := r.Get(d.Doc.ClusterID, d.Doc.ID, d.Doc.AccessHash)
            if err != nil {
                return false
            } else if _, err = os.Stat(r.GetFilePath(f)); os.IsNotExist(err) {
                return false
            }

            switch msg.PeerType(m.PeerType) {
            case msg.PeerType_PeerUser:
                userMtx.Lock()
                if _, ok := userMediaInfo[m.PeerID]; !ok {
                    userMediaInfo[m.PeerID] = make(map[msg.ClientMediaType]int64, 5)
                }
                userMediaInfo[m.PeerID][msg.ClientMediaType(item.UserMeta())] += int64(d.Doc.FileSize)
                userMtx.Unlock()
            case msg.PeerType_PeerGroup:
                groupMtx.Lock()
                if _, ok := groupMediaInfo[m.PeerID]; !ok {
                    groupMediaInfo[m.PeerID] = make(map[msg.ClientMediaType]int64, 5)
                }
                groupMediaInfo[m.PeerID][msg.ClientMediaType(item.UserMeta())] += int64(d.Doc.FileSize)

                groupMtx.Unlock()
            }
        default:
            return false
        }
        return true
    }
    stream.Send = func(list *pb.KVList) error {
        return nil
    }

    _ = stream.Orchestrate(context.Background())

    cachedMediaInfo := &msg.ClientCachedMediaInfo{}
    for peerID, mi := range userMediaInfo {
        peerInfo := &msg.ClientPeerMediaInfo{
            PeerID:   peerID,
            PeerType: msg.PeerType_PeerUser,
        }
        for mType, mSize := range mi {
            peerInfo.Media = append(peerInfo.Media, &msg.ClientMediaSize{
                MediaType: mType,
                TotalSize: mSize,
            })
        }
        cachedMediaInfo.MediaInfo = append(cachedMediaInfo.MediaInfo, peerInfo)
    }
    for peerID, mi := range groupMediaInfo {
        peerInfo := &msg.ClientPeerMediaInfo{
            PeerID:   peerID,
            PeerType: msg.PeerType_PeerGroup,
        }
        for mType, mSize := range mi {
            peerInfo.Media = append(peerInfo.Media, &msg.ClientMediaSize{
                MediaType: mType,
                TotalSize: mSize,
            })
        }
        cachedMediaInfo.MediaInfo = append(cachedMediaInfo.MediaInfo, peerInfo)
    }
    return cachedMediaInfo
}

func (r *repoFiles) DeleteCachedMediaByPeer(teamID, peerID int64, peerType int32, mediaTypes []msg.ClientMediaType) {
    stream := r.badger.NewStream()

    stream.Prefix = getMessagePrefix(teamID, peerID, peerType)
    stream.ChooseKey = func(item *badger.Item) bool {
        m := &msg.UserMessage{}
        err := item.Value(func(val []byte) error {
            return m.Unmarshal(val)
        })
        if err != nil {
            return false
        }
        switch m.MediaType {
        case msg.MediaType_MediaTypeDocument:
            d := msg.MediaDocument{}
            err = d.Unmarshal(m.Media)
            if err != nil {
                return false
            }
            for _, mt := range mediaTypes {
                if msg.ClientMediaType(item.UserMeta()) == mt {
                    f, err := r.Get(d.Doc.ClusterID, d.Doc.ID, d.Doc.AccessHash)
                    if err != nil {
                        return false
                    }
                    _ = os.Remove(r.GetFilePath(f))
                    return true
                }
            }

        default:
            return false
        }
        return true
    }
    stream.Send = func(list *pb.KVList) error {
        return nil
    }

    _ = stream.Orchestrate(context.Background())
}

func (r *repoFiles) DeleteCachedMediaByMediaType(teamID int64, mediaTypes []msg.ClientMediaType) {
    stream := r.badger.NewStream()

    stream.Prefix = []byte(fmt.Sprintf("%s.", prefixMessages))
    stream.ChooseKey = func(item *badger.Item) bool {
        m := &msg.UserMessage{}
        err := item.Value(func(val []byte) error {
            return m.Unmarshal(val)
        })
        if err != nil {
            return false
        }
        if m.TeamID != teamID {
            return false
        }
        switch m.MediaType {
        case msg.MediaType_MediaTypeDocument:
            d := msg.MediaDocument{}
            err = d.Unmarshal(m.Media)
            if err != nil {
                return false
            }
            for _, mt := range mediaTypes {
                if msg.ClientMediaType(item.UserMeta()) == mt {
                    f, err := r.Get(d.Doc.ClusterID, d.Doc.ID, d.Doc.AccessHash)
                    if err != nil {
                        return false
                    }
                    _ = os.Remove(r.GetFilePath(f))
                    return true
                }
            }

        default:
            return false
        }
        return true
    }
    stream.Send = func(list *pb.KVList) error {
        return nil
    }

    _ = stream.Orchestrate(context.Background())
}

func (r *repoFiles) ClearCache() {
    dirs := []string{
        DirAudio, DirFile, DirPhoto, DirVideo, DirCache,
    }
    for _, dir := range dirs {
        _ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
            if info.IsDir() {
                return nil
            }
            _ = os.Remove(path)
            return nil
        })
    }
}

func (r *repoFiles) GetFilePath(clientFile *msg.ClientFile) string {
    switch clientFile.Type {
    case msg.ClientFileType_Gif:
        fallthrough
    case msg.ClientFileType_Message:
        return getMessageFilePath(clientFile.MimeType, clientFile.FileID, clientFile.Extension)
    case msg.ClientFileType_AccountProfilePhoto:
        return getAccountProfilePath(clientFile.UserID, clientFile.FileID)
    case msg.ClientFileType_GroupProfilePhoto:
        return getGroupProfilePath(clientFile.GroupID, clientFile.FileID)
    case msg.ClientFileType_Thumbnail:
        return getThumbnailPath(clientFile.FileID, clientFile.ClusterID)
    case msg.ClientFileType_Wallpaper:
        return getWallpaperPath(clientFile.FileID, clientFile.ClusterID)
    }
    return ""
}

func getMessageFilePath(mimeType string, docID int64, ext string) string {
    mimeType = strings.ToLower(mimeType)
    if ext == "" {
        exts, _ := mime.ExtensionsByType(mimeType)
        if len(exts) > 0 {
            ext = exts[len(exts)-1]
        }
    }

    // if the file is opus type,
    // means its voice file so it should be saved in cache folder
    // so user could not access to it by file manager
    switch {
    case mimeType == "audio/ogg":
        ext = ".ogg"
        return path.Join(DirCache, fmt.Sprintf("%d%s", docID, ext))
    case strings.HasPrefix(mimeType, "video/"):
        return path.Join(DirVideo, fmt.Sprintf("%d%s", docID, ext))
    case strings.HasPrefix(mimeType, "audio/"):
        return path.Join(DirAudio, fmt.Sprintf("%d%s", docID, ext))
    case strings.HasPrefix(mimeType, "image/"):
        return path.Join(DirPhoto, fmt.Sprintf("%d%s", docID, ext))
    default:
        return path.Join(DirFile, fmt.Sprintf("%d%s", docID, ext))
    }
}

func getThumbnailPath(fileID int64, clusterID int32) string {
    return path.Join(DirCache, fmt.Sprintf("%d%d%s", fileID, clusterID, ".jpg"))
}

func getWallpaperPath(fileID int64, clusterID int32) string {
    return path.Join(DirPhoto, fmt.Sprintf("%s_%d%d%s", "Wallpaper", fileID, clusterID, ".jpg"))
}

func getAccountProfilePath(userID int64, fileID int64) string {
    return path.Join(DirCache, fmt.Sprintf("u%d_%d%s", userID, fileID, ".jpg"))
}

func getGroupProfilePath(groupID int64, fileID int64) string {
    return path.Join(DirCache, fmt.Sprintf("g%d_%d%s", groupID, fileID, ".jpg"))
}

func (r *repoFiles) SaveFileRequest(reqID string, req *msg.ClientFileRequest, overwriteOnly bool) (bool, error) {
    var saved bool
    err := badgerUpdate(func(txn *badger.Txn) error {
        key := tools.StrToByte(fmt.Sprintf("%s.%s", prefixFilesRequests, reqID))
        if overwriteOnly {
            _, err := txn.Get(key)
            if err != nil {
                return nil
            }
        }
        reqBytes, _ := req.Marshal()
        err := txn.Set(key, reqBytes)
        if err != nil {
            return err
        }
        saved = true
        return nil
    })
    return saved, err
}

func (r *repoFiles) DeleteFileRequest(reqID string) error {
    return badgerUpdate(func(txn *badger.Txn) error {
        return txn.Delete(
            tools.StrToByte(fmt.Sprintf("%s.%s", prefixFilesRequests, reqID)),
        )
    })
}

func (r *repoFiles) GetFileRequest(reqID string) (*msg.ClientFileRequest, error) {
    req := &msg.ClientFileRequest{}
    err := badgerView(func(txn *badger.Txn) error {
        item, err := txn.Get(
            tools.StrToByte(fmt.Sprintf("%s.%s", prefixFilesRequests, reqID)),
        )
        switch err {
        case nil:
        case badger.ErrKeyNotFound:
            return domain.ErrNotFound
        default:
            return err

        }
        return item.Value(func(val []byte) error {
            return req.Unmarshal(val)
        })
    })
    return req, err
}

func (r *repoFiles) GetAllFileRequests() ([]*msg.ClientFileRequest, error) {
    reqs := make([]*msg.ClientFileRequest, 0, 8)
    st := r.badger.NewStream()
    st.Prefix = tools.StrToByte(prefixFilesRequests)
    st.ChooseKey = func(item *badger.Item) bool {
        return true
    }
    st.Send = func(list *badger.KVList) error {
        for _, kv := range list.Kv {
            req := &msg.ClientFileRequest{}
            err := req.Unmarshal(kv.Value)
            if err != nil {
                return err
            }
            reqs = append(reqs, req)
        }
        return nil
    }
    err := st.Orchestrate(context.Background())
    if err != nil {
        return nil, err
    }
    return reqs, nil
}
