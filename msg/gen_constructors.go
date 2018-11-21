package msg

const (
	C_MessagesSendMedia            int64 = 25498545
	C_Peer                         int64 = 47470215
	C_MessagesEditGroupTitle       int64 = 143417640
	C_UpdateUserTyping             int64 = 178254060
	C_Ack                          int64 = 447331921
	C_ContactUser                  int64 = 460099170
	C_AccountGetNotifySettings     int64 = 477008681
	C_UpdateReadHistoryOutbox      int64 = 510866108
	C_MessageEnvelope              int64 = 535232465
	C_UpdateGetDifference          int64 = 556775761
	C_InitAuthCompleted            int64 = 627708982
	C_UpdateContainer              int64 = 661712615
	C_File                         int64 = 749574446
	C_User                         int64 = 765557111
	C_UsersMany                    int64 = 801733941
	C_AccountRegisterDevice        int64 = 946059841
	C_AuthLogout                   int64 = 992431648
	C_UsersGet                     int64 = 1039301579
	C_RSAPublicKey                 int64 = 1046601890
	C_MessagesGetDialog            int64 = 1050840034
	C_Dialog                       int64 = 1120787796
	C_AuthAuthorization            int64 = 1140037965
	C_AuthRecall                   int64 = 1172029049
	C_SystemGetPublicKeys          int64 = 1191522796
	C_MessagesReadHistory          int64 = 1300826534
	C_ContactsGet                  int64 = 1412732665
	C_MessagesGetDialogs           int64 = 1429532372
	C_UpdateGetState               int64 = 1437250230
	C_AccountUpdateUsername        int64 = 1477164344
	C_AccountCheckUsername         int64 = 1501406413
	C_UpdateReadHistoryInbox       int64 = 1529128378
	C_MessagesSetTyping            int64 = 1540214486
	C_InitCompleteAuth             int64 = 1583178320
	C_UserMessage                  int64 = 1677556362
	C_MessagesMany                 int64 = 1713238910
	C_UpdateDifference             int64 = 1742546619
	C_ContactsDelete               int64 = 1750426880
	C_SystemGetDHGroups            int64 = 1786665018
	C_UpdateMessageEdited          int64 = 1825079988
	C_UpdateState                  int64 = 1837585836
	C_MessageContainer             int64 = 1972016308
	C_AccountSetNotifySettings     int64 = 2016882075
	C_UpdateMessageID              int64 = 2139063022
	C_MessagesGet                  int64 = 2151382317
	C_ContactsImported             int64 = 2157298354
	C_ClientPendingMessage         int64 = 2164891929
	C_ProtoMessage                 int64 = 2179260159
	C_AuthRegister                 int64 = 2228369460
	C_AuthCheckedPhone             int64 = 2236203131
	C_SystemClientLog              int64 = 2244397803
	C_InitCompleteAuthInternal     int64 = 2360982492
	C_UpdateEnvelope               int64 = 2373884514
	C_AuthSentCode                 int64 = 2375498471
	C_MessagesDeleteGroupUser      int64 = 2410057319
	C_FileLocation                 int64 = 2432133155
	C_MessagesEdit                 int64 = 2492658432
	C_MessagesAddGroupUser         int64 = 2559309146
	C_AuthLogin                    int64 = 2587620888
	C_Error                        int64 = 2619118453
	C_ProtoEncryptedPayload        int64 = 2668405547
	C_PhoneContact                 int64 = 2672574672
	C_UpdateUserStatus             int64 = 2696747995
	C_SystemPublicKeys             int64 = 2745130223
	C_DHGroup                      int64 = 2751503049
	C_InitDB                       int64 = 2793857427
	C_UpdateChatMemberDeleted      int64 = 2798925845
	C_EchoWithDelay                int64 = 2861516000
	C_Group                        int64 = 2885774273
	C_SystemDHGroups               int64 = 2890748083
	C_MessagesCreateGroup          int64 = 2916726768
	C_MessagesSent                 int64 = 2942502835
	C_MessagesSend                 int64 = 3000244183
	C_UpdateChatMemberAdded        int64 = 3034247697
	C_UpdateNotifySettings         int64 = 3187524885
	C_AuthRecalled                 int64 = 3249025459
	C_MessagesDialogs              int64 = 3252610224
	C_InputPeer                    int64 = 3374092470
	C_MessagesGetHistory           int64 = 3396939832
	C_UpdateNewMessage             int64 = 3426925183
	C_ContactsImport               int64 = 3473528730
	C_PeerNotifySettings           int64 = 3475030132
	C_AccountUpdateProfile         int64 = 3725499887
	C_FileSavePart                 int64 = 3766876582
	C_InputUser                    int64 = 3865689926
	C_ClientPendingMessageDelivery int64 = 3881219190
	C_InputFile                    int64 = 3882180383
	C_ContactsMany                 int64 = 3883395672
	C_AccountUnregisterDevice      int64 = 3981251588
	C_AuthSendCode                 int64 = 3984043365
	C_Bool                         int64 = 4122188204
	C_InitResponse                 int64 = 4130340247
	C_AuthCheckPhone               int64 = 4134648516
	C_InitConnect                  int64 = 4150793517
	C_FileGet                      int64 = 4282510672
	C_UpdateUsername               int64 = 4290110589
)

var ConstructorNames = map[int64]string{
	25498545:   "MessagesSendMedia",
	47470215:   "Peer",
	143417640:  "MessagesEditGroupTitle",
	178254060:  "UpdateUserTyping",
	447331921:  "Ack",
	460099170:  "ContactUser",
	477008681:  "AccountGetNotifySettings",
	510866108:  "UpdateReadHistoryOutbox",
	535232465:  "MessageEnvelope",
	556775761:  "UpdateGetDifference",
	627708982:  "InitAuthCompleted",
	661712615:  "UpdateContainer",
	749574446:  "File",
	765557111:  "User",
	801733941:  "UsersMany",
	946059841:  "AccountRegisterDevice",
	992431648:  "AuthLogout",
	1039301579: "UsersGet",
	1046601890: "RSAPublicKey",
	1050840034: "MessagesGetDialog",
	1120787796: "Dialog",
	1140037965: "AuthAuthorization",
	1172029049: "AuthRecall",
	1191522796: "SystemGetPublicKeys",
	1300826534: "MessagesReadHistory",
	1412732665: "ContactsGet",
	1429532372: "MessagesGetDialogs",
	1437250230: "UpdateGetState",
	1477164344: "AccountUpdateUsername",
	1501406413: "AccountCheckUsername",
	1529128378: "UpdateReadHistoryInbox",
	1540214486: "MessagesSetTyping",
	1583178320: "InitCompleteAuth",
	1677556362: "UserMessage",
	1713238910: "MessagesMany",
	1742546619: "UpdateDifference",
	1750426880: "ContactsDelete",
	1786665018: "SystemGetDHGroups",
	1825079988: "UpdateMessageEdited",
	1837585836: "UpdateState",
	1972016308: "MessageContainer",
	2016882075: "AccountSetNotifySettings",
	2139063022: "UpdateMessageID",
	2151382317: "MessagesGet",
	2157298354: "ContactsImported",
	2164891929: "ClientPendingMessage",
	2179260159: "ProtoMessage",
	2228369460: "AuthRegister",
	2236203131: "AuthCheckedPhone",
	2244397803: "SystemClientLog",
	2360982492: "InitCompleteAuthInternal",
	2373884514: "UpdateEnvelope",
	2375498471: "AuthSentCode",
	2410057319: "MessagesDeleteGroupUser",
	2432133155: "FileLocation",
	2492658432: "MessagesEdit",
	2559309146: "MessagesAddGroupUser",
	2587620888: "AuthLogin",
	2619118453: "Error",
	2668405547: "ProtoEncryptedPayload",
	2672574672: "PhoneContact",
	2696747995: "UpdateUserStatus",
	2745130223: "SystemPublicKeys",
	2751503049: "DHGroup",
	2793857427: "InitDB",
	2798925845: "UpdateChatMemberDeleted",
	2861516000: "EchoWithDelay",
	2885774273: "Group",
	2890748083: "SystemDHGroups",
	2916726768: "MessagesCreateGroup",
	2942502835: "MessagesSent",
	3000244183: "MessagesSend",
	3034247697: "UpdateChatMemberAdded",
	3187524885: "UpdateNotifySettings",
	3249025459: "AuthRecalled",
	3252610224: "MessagesDialogs",
	3374092470: "InputPeer",
	3396939832: "MessagesGetHistory",
	3426925183: "UpdateNewMessage",
	3473528730: "ContactsImport",
	3475030132: "PeerNotifySettings",
	3725499887: "AccountUpdateProfile",
	3766876582: "FileSavePart",
	3865689926: "InputUser",
	3881219190: "ClientPendingMessageDelivery",
	3882180383: "InputFile",
	3883395672: "ContactsMany",
	3981251588: "AccountUnregisterDevice",
	3984043365: "AuthSendCode",
	4122188204: "Bool",
	4130340247: "InitResponse",
	4134648516: "AuthCheckPhone",
	4150793517: "InitConnect",
	4282510672: "FileGet",
	4290110589: "UpdateUsername",
}
