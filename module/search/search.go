package search

import (
    "github.com/ronaksoft/river-msg/go/msg"
    "github.com/ronaksoft/river-sdk/internal/request"
    "github.com/ronaksoft/river-sdk/module"
)

/*
   Creation Time: 2021 - May - 10
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

type search struct {
    module.Base
}

func New() *search {
    r := &search{}
    r.RegisterHandlers(
        map[int64]request.LocalHandler{
            msg.C_ClientGetRecentSearch:         r.clientGetRecentSearch,
            msg.C_ClientGlobalSearch:            r.clientGlobalSearch,
            msg.C_ClientPutRecentSearch:         r.clientPutRecentSearch,
            msg.C_ClientRemoveAllRecentSearches: r.clientRemoveAllRecentSearches,
            msg.C_ClientRemoveRecentSearch:      r.clientRemoveRecentSearch,
        },
    )
    return r
}

func (r *search) Name() string {
    return module.Search
}
