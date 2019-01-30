package msg

const (
	C_MessagesSendMedia                  int64 = 25498545
	C_Peer                               int64 = 47470215
	C_MediaPhoto                         int64 = 71894788
	C_InitUserBound                      int64 = 128391141
	C_MediaWebPage                       int64 = 148034084
	C_UpdateUserTyping                   int64 = 178254060
	C_InputMediaGeoLocation              int64 = 185664060
	C_GroupFull                          int64 = 205850814
	C_UpdateUserPhoto                    int64 = 302028082
	C_DocumentAttributeAudio             int64 = 309707708
	C_AccountPrivacyDisallowUsers        int64 = 352774342
	C_InputFileLocation                  int64 = 354669666
	C_UpdateGroupPhoto                   int64 = 367193154
	C_GroupsAddUser                      int64 = 394654713
	C_AccountUpdatePhoto                 int64 = 406174115
	C_Ack                                int64 = 447331921
	C_ContactUser                        int64 = 460099170
	C_AccountGetNotifySettings           int64 = 477008681
	C_UpdateReadHistoryOutbox            int64 = 510866108
	C_DocumentAttributePhoto             int64 = 515862833
	C_MessageEnvelope                    int64 = 535232465
	C_Document                           int64 = 555739168
	C_UpdateGetDifference                int64 = 556775761
	C_InitAuthCompleted                  int64 = 627708982
	C_AccountPrivacyAllowAll             int64 = 642273673
	C_UpdateContainer                    int64 = 661712615
	C_UpdateMessagesDeleted              int64 = 670568714
	C_AccountPrivacyRule                 int64 = 690047132
	C_UpdateGroupAdmins                  int64 = 694155405
	C_GroupUploadPhoto                   int64 = 703740190
	C_File                               int64 = 749574446
	C_User                               int64 = 765557111
	C_UsersMany                          int64 = 801733941
	C_InputMediaUploadedPhoto            int64 = 849930963
	C_InputMediaUploadedDocument         int64 = 870692909
	C_AccountRegisterDevice              int64 = 946059841
	C_AuthLogout                         int64 = 992431648
	C_UsersGet                           int64 = 1039301579
	C_RSAPublicKey                       int64 = 1046601890
	C_MessagesGetDialog                  int64 = 1050840034
	C_ClientSendMessageMedia             int64 = 1095038539
	C_Dialog                             int64 = 1120787796
	C_AuthAuthorization                  int64 = 1140037965
	C_AuthRecall                         int64 = 1172029049
	C_SystemGetPublicKeys                int64 = 1191522796
	C_MessageActionGroupDeleteUser       int64 = 1213452128
	C_AccountUploadPhoto                 int64 = 1222469957
	C_MessageActionClearHistory          int64 = 1270465696
	C_GroupsCreate                       int64 = 1271969037
	C_MessagesReadHistory                int64 = 1300826534
	C_GroupsUpdateAdmin                  int64 = 1345991011
	C_ContactsGet                        int64 = 1412732665
	C_MessagesGetDialogs                 int64 = 1429532372
	C_UpdateGetState                     int64 = 1437250230
	C_AccountUpdateUsername              int64 = 1477164344
	C_AccountCheckUsername               int64 = 1501406413
	C_UpdateReadHistoryInbox             int64 = 1529128378
	C_InputMediaContact                  int64 = 1534117184
	C_MessagesSetTyping                  int64 = 1540214486
	C_GroupsToggleAdmins                 int64 = 1581076909
	C_InitCompleteAuth                   int64 = 1583178320
	C_AccountSetPrivacy                  int64 = 1599585002
	C_UpdateGroupParticipantAdd          int64 = 1623827837
	C_UserMessage                        int64 = 1677556362
	C_MessagesMany                       int64 = 1713238910
	C_UpdateDifference                   int64 = 1742546619
	C_ContactsDelete                     int64 = 1750426880
	C_MessagesReadContents               int64 = 1781251275
	C_SystemGetDHGroups                  int64 = 1786665018
	C_UpdateGroupParticipantAdmin        int64 = 1813022164
	C_UpdateMessageEdited                int64 = 1825079988
	C_UpdateState                        int64 = 1837585836
	C_UserPhoto                          int64 = 1881347437
	C_AccountGetPrivacy                  int64 = 1897044856
	C_InitBindUser                       int64 = 1933549113
	C_MessageActionGroupAddUser          int64 = 1949386261
	C_MessageContainer                   int64 = 1972016308
	C_MessagesClearHistory               int64 = 1981246180
	C_DocumentAttributeVideo             int64 = 1993289477
	C_AccountSetNotifySettings           int64 = 2016882075
	C_UpdateMessageID                    int64 = 2139063022
	C_MessagesGet                        int64 = 2151382317
	C_ContactsImported                   int64 = 2157298354
	C_ClientPendingMessage               int64 = 2164891929
	C_ProtoMessage                       int64 = 2179260159
	C_InputMediaPhoto                    int64 = 2201579839
	C_DocumentAttributeFile              int64 = 2227452062
	C_AuthRegister                       int64 = 2228369460
	C_AuthCheckedPhone                   int64 = 2236203131
	C_MessageActionGroupCreated          int64 = 2241024808
	C_InputMediaDocument                 int64 = 2258657627
	C_MediaDocument                      int64 = 2281620705
	C_InitCompleteAuthInternal           int64 = 2360982492
	C_UpdateEnvelope                     int64 = 2373884514
	C_AuthSentCode                       int64 = 2375498471
	C_MessageActionContactRegistered     int64 = 2399156016
	C_MessageActionGroupTitleChanged     int64 = 2418464749
	C_FileLocation                       int64 = 2432133155
	C_UpdateGroupParticipantDeleted      int64 = 2489941844
	C_MessagesEdit                       int64 = 2492658432
	C_GroupsEditTitle                    int64 = 2582813461
	C_AuthLogin                          int64 = 2587620888
	C_Error                              int64 = 2619118453
	C_MediaGeoLocation                   int64 = 2625326500
	C_MessagesForward                    int64 = 2662884753
	C_ProtoEncryptedPayload              int64 = 2668405547
	C_PhoneContact                       int64 = 2672574672
	C_AuthResendCode                     int64 = 2682713491
	C_UpdateUserStatus                   int64 = 2696747995
	C_SystemPublicKeys                   int64 = 2745130223
	C_DHGroup                            int64 = 2751503049
	C_InitDB                             int64 = 2793857427
	C_AuthLoginByToken                   int64 = 2851553023
	C_EchoWithDelay                      int64 = 2861516000
	C_Group                              int64 = 2885774273
	C_SystemDHGroups                     int64 = 2890748083
	C_MessagesSent                       int64 = 2942502835
	C_GroupsGetFull                      int64 = 2986704909
	C_UpdateReadMessagesContents         int64 = 2991403048
	C_MessagesSend                       int64 = 3000244183
	C_ClientUpdateMessagesDeleted        int64 = 3060926862
	C_GroupsDeleteUser                   int64 = 3172322223
	C_UpdateNotifySettings               int64 = 3187524885
	C_AuthRecalled                       int64 = 3249025459
	C_MessagesDialogs                    int64 = 3252610224
	C_UsersGetFull                       int64 = 3343342086
	C_InputPeer                          int64 = 3374092470
	C_MessagesGetHistory                 int64 = 3396939832
	C_UpdateNewMessage                   int64 = 3426925183
	C_AccountPrivacyAllowContacts        int64 = 3442577898
	C_ContactsImport                     int64 = 3473528730
	C_PeerNotifySettings                 int64 = 3475030132
	C_MessageEntity                      int64 = 3479443932
	C_MessagesDelete                     int64 = 3487616910
	C_AccountPrivacyDisallowContacts     int64 = 3629404744
	C_AccountPrivacyDisallowAll          int64 = 3653779109
	C_AuthDestroyKey                     int64 = 3673422656
	C_AccountUpdateProfile               int64 = 3725499887
	C_AccountRemovePhoto                 int64 = 3728692172
	C_MediaContact                       int64 = 3735320833
	C_FileSavePart                       int64 = 3766876582
	C_AccountPrivacyRules                int64 = 3802018092
	C_ClientUpdatePendingMessageDelivery int64 = 3828722061
	C_InputUser                          int64 = 3865689926
	C_InputFile                          int64 = 3882180383
	C_ContactsMany                       int64 = 3883395672
	C_AccountUnregisterDevice            int64 = 3981251588
	C_AuthSendCode                       int64 = 3984043365
	C_GroupPhoto                         int64 = 3998516135
	C_AccountPrivacyAllowUsers           int64 = 4047110177
	C_GroupParticipant                   int64 = 4072279665
	C_InputDocument                      int64 = 4081048424
	C_Bool                               int64 = 4122188204
	C_InitResponse                       int64 = 4130340247
	C_AuthCheckPhone                     int64 = 4134648516
	C_DocumentAttribute                  int64 = 4146719643
	C_InitConnect                        int64 = 4150793517
	C_FileGet                            int64 = 4282510672
	C_AccountChangePhone                 int64 = 4285969474
	C_UpdateUsername                     int64 = 4290110589
)

var ConstructorNames = map[int64]string{
	25498545:   "MessagesSendMedia",
	47470215:   "Peer",
	71894788:   "MediaPhoto",
	128391141:  "InitUserBound",
	148034084:  "MediaWebPage",
	178254060:  "UpdateUserTyping",
	185664060:  "InputMediaGeoLocation",
	205850814:  "GroupFull",
	302028082:  "UpdateUserPhoto",
	309707708:  "DocumentAttributeAudio",
	352774342:  "AccountPrivacyDisallowUsers",
	354669666:  "InputFileLocation",
	367193154:  "UpdateGroupPhoto",
	394654713:  "GroupsAddUser",
	406174115:  "AccountUpdatePhoto",
	447331921:  "Ack",
	460099170:  "ContactUser",
	477008681:  "AccountGetNotifySettings",
	510866108:  "UpdateReadHistoryOutbox",
	515862833:  "DocumentAttributePhoto",
	535232465:  "MessageEnvelope",
	555739168:  "Document",
	556775761:  "UpdateGetDifference",
	627708982:  "InitAuthCompleted",
	642273673:  "AccountPrivacyAllowAll",
	661712615:  "UpdateContainer",
	670568714:  "UpdateMessagesDeleted",
	690047132:  "AccountPrivacyRule",
	694155405:  "UpdateGroupAdmins",
	703740190:  "GroupUploadPhoto",
	749574446:  "File",
	765557111:  "User",
	801733941:  "UsersMany",
	849930963:  "InputMediaUploadedPhoto",
	870692909:  "InputMediaUploadedDocument",
	946059841:  "AccountRegisterDevice",
	992431648:  "AuthLogout",
	1039301579: "UsersGet",
	1046601890: "RSAPublicKey",
	1050840034: "MessagesGetDialog",
	1095038539: "ClientSendMessageMedia",
	1120787796: "Dialog",
	1140037965: "AuthAuthorization",
	1172029049: "AuthRecall",
	1191522796: "SystemGetPublicKeys",
	1213452128: "MessageActionGroupDeleteUser",
	1222469957: "AccountUploadPhoto",
	1270465696: "MessageActionClearHistory",
	1271969037: "GroupsCreate",
	1300826534: "MessagesReadHistory",
	1345991011: "GroupsUpdateAdmin",
	1412732665: "ContactsGet",
	1429532372: "MessagesGetDialogs",
	1437250230: "UpdateGetState",
	1477164344: "AccountUpdateUsername",
	1501406413: "AccountCheckUsername",
	1529128378: "UpdateReadHistoryInbox",
	1534117184: "InputMediaContact",
	1540214486: "MessagesSetTyping",
	1581076909: "GroupsToggleAdmins",
	1583178320: "InitCompleteAuth",
	1599585002: "AccountSetPrivacy",
	1623827837: "UpdateGroupParticipantAdd",
	1677556362: "UserMessage",
	1713238910: "MessagesMany",
	1742546619: "UpdateDifference",
	1750426880: "ContactsDelete",
	1781251275: "MessagesReadContents",
	1786665018: "SystemGetDHGroups",
	1813022164: "UpdateGroupParticipantAdmin",
	1825079988: "UpdateMessageEdited",
	1837585836: "UpdateState",
	1881347437: "UserPhoto",
	1897044856: "AccountGetPrivacy",
	1933549113: "InitBindUser",
	1949386261: "MessageActionGroupAddUser",
	1972016308: "MessageContainer",
	1981246180: "MessagesClearHistory",
	1993289477: "DocumentAttributeVideo",
	2016882075: "AccountSetNotifySettings",
	2139063022: "UpdateMessageID",
	2151382317: "MessagesGet",
	2157298354: "ContactsImported",
	2164891929: "ClientPendingMessage",
	2179260159: "ProtoMessage",
	2201579839: "InputMediaPhoto",
	2227452062: "DocumentAttributeFile",
	2228369460: "AuthRegister",
	2236203131: "AuthCheckedPhone",
	2241024808: "MessageActionGroupCreated",
	2258657627: "InputMediaDocument",
	2281620705: "MediaDocument",
	2360982492: "InitCompleteAuthInternal",
	2373884514: "UpdateEnvelope",
	2375498471: "AuthSentCode",
	2399156016: "MessageActionContactRegistered",
	2418464749: "MessageActionGroupTitleChanged",
	2432133155: "FileLocation",
	2489941844: "UpdateGroupParticipantDeleted",
	2492658432: "MessagesEdit",
	2582813461: "GroupsEditTitle",
	2587620888: "AuthLogin",
	2619118453: "Error",
	2625326500: "MediaGeoLocation",
	2662884753: "MessagesForward",
	2668405547: "ProtoEncryptedPayload",
	2672574672: "PhoneContact",
	2682713491: "AuthResendCode",
	2696747995: "UpdateUserStatus",
	2745130223: "SystemPublicKeys",
	2751503049: "DHGroup",
	2793857427: "InitDB",
	2851553023: "AuthLoginByToken",
	2861516000: "EchoWithDelay",
	2885774273: "Group",
	2890748083: "SystemDHGroups",
	2942502835: "MessagesSent",
	2986704909: "GroupsGetFull",
	2991403048: "UpdateReadMessagesContents",
	3000244183: "MessagesSend",
	3060926862: "ClientUpdateMessagesDeleted",
	3172322223: "GroupsDeleteUser",
	3187524885: "UpdateNotifySettings",
	3249025459: "AuthRecalled",
	3252610224: "MessagesDialogs",
	3343342086: "UsersGetFull",
	3374092470: "InputPeer",
	3396939832: "MessagesGetHistory",
	3426925183: "UpdateNewMessage",
	3442577898: "AccountPrivacyAllowContacts",
	3473528730: "ContactsImport",
	3475030132: "PeerNotifySettings",
	3479443932: "MessageEntity",
	3487616910: "MessagesDelete",
	3629404744: "AccountPrivacyDisallowContacts",
	3653779109: "AccountPrivacyDisallowAll",
	3673422656: "AuthDestroyKey",
	3725499887: "AccountUpdateProfile",
	3728692172: "AccountRemovePhoto",
	3735320833: "MediaContact",
	3766876582: "FileSavePart",
	3802018092: "AccountPrivacyRules",
	3828722061: "ClientUpdatePendingMessageDelivery",
	3865689926: "InputUser",
	3882180383: "InputFile",
	3883395672: "ContactsMany",
	3981251588: "AccountUnregisterDevice",
	3984043365: "AuthSendCode",
	3998516135: "GroupPhoto",
	4047110177: "AccountPrivacyAllowUsers",
	4072279665: "GroupParticipant",
	4081048424: "InputDocument",
	4122188204: "Bool",
	4130340247: "InitResponse",
	4134648516: "AuthCheckPhone",
	4146719643: "DocumentAttribute",
	4150793517: "InitConnect",
	4282510672: "FileGet",
	4285969474: "AccountChangePhone",
	4290110589: "UpdateUsername",
}
