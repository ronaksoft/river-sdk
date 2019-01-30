package msg

import (
    "errors"
)

/*
   Creation Time: 2018 - Apr - 07
   Created by:  Ehsan N. Moosa (ehsan)
   Maintainers:
       1.  Ehsan N. Moosa (ehsan)
   Auditor: Ehsan N. Moosa
   Copyright Ronak Software Group 2018
*/

var (
    ErrNoHandler     = errors.New("no handler submitted")
    ErrTimeout       = errors.New("time out")
    ErrIncorrectSize = errors.New("incorrect size")
)

// Error Codes
const (
    ERR_CODE_INTERNAL       = "E00"
    ERR_CODE_INVALID        = "E01"
    ERR_CODE_UNAVAILABLE    = "E02"
    ERR_CODE_TOO_MANY       = "E03"
    ERR_CODE_TOO_FEW        = "E04"
    ERR_CODE_INCOMPLETE     = "E05"
    ERR_CODE_TIMEOUT        = "E06"
    ERR_CODE_ACCESS         = "E07"
    ERR_CODE_ALREADY_EXISTS = "E08"
    ERR_CODE_BUSY           = "E09"
    ERR_CODE_OUT_OF_RANGE   = "E10"
)

// Error Items
const (
    ERR_ITEM_PHONE          = "PHONE"
    ERR_ITEM_PHONE_CODE     = "PHONE_CODE"
    ERR_ITEM_USER_ID        = "USER_ID"
    ERR_ITEM_PEER           = "PEER"
    ERR_ITEM_PEER_TYPE      = "PEER_TYPE"
    ERR_ITEM_INPUT          = "INPUT"
    ERR_ITEM_REQUEST        = "REQUEST"
    ERR_ITEM_MESSAGE        = "MESSAGE"
    ERR_ITEM_MESSAGE_ID     = "MESSAGE_ID"
    ERR_ITEM_SERVER         = "SERVER"
    ERR_ITEM_PQ             = "PQ"
    ERR_ITEM_ENCRYPTION     = "ENCRYPTION"
    ERR_ITEM_RSA_KEY        = "RSA_KEY"
    ERR_ITEM_PROTO          = "PROTO"
    ERR_ITEM_DH_KEY         = "DH_KEY"
    ERR_ITEM_SIGN_IN        = "SIGN_IN"
    ERR_ITEM_RANDOM_ID      = "RANDOM_ID"
    ERR_ITEM_ACCESS_HASH    = "ACCESS_HASH"
    ERR_ITEM_JOB_WORKER     = "JOB_WORKER"
    ERR_ITEM_AUTH           = "AUTH"
    ERR_ITEM_USERNAME       = "USERNAME"
    ERR_ITEM_CHAT_TEXT      = "CHAT_TEXT"
    ERR_ITEM_GROUP_TITLE    = "GROUP_TITLE"
    ERR_ITEM_GROUP_ID       = "GROUP_ID"
    ERR_ITEM_GROUP          = "GROUP"
    ERR_ITEM_USERS          = "USERS"
    ERR_ITEM_RETRACT_TIME   = "RETRACT_TIME"
    ERR_ITEM_BIO            = "BIO"
    ERR_ITEM_API            = "API"
    ERR_ITEM_LAST_ADMIN     = "LAST_ADMIN"
    ERR_ITEM_DELETE_CREATOR = "DELETE_CREATOR"
    ERR_ITEM_FILE_PART_ID   = "FILE_PART_ID"
    ERR_ITEM_FILE_PARTS     = "FILE_PARTS"
    ERR_ITEM_FILE_PART_SIZE = "FILE_PART_SIZE"
    ERR_ITEM_DEVICE_TOKEN   = "DEVICE_TOKEN"
    ERR_ITEM_DEVICE_MODEL   = "DEVICE_MODEL"
    ERR_ITEM_DOCUMENT       = "DOCUMENT"
    ERR_ITEM_TOKEN          = "TOKEN"
)
