package request

/*
   Creation Time: 2021 - May - 25
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2020
*/

// Status state of file download/upload request
type Status int32

const (
	StatusNone       Status = 0 // RequestStatusNone no request invoked
	StatusInProgress Status = 1 // RequestStatusInProgress downloading/uploading
	StatusCompleted  Status = 2 // RequestStatusCompleted already file is downloaded/uploaded
	StatusCanceled   Status = 4 // RequestStatusCanceled canceled by user
	StatusError      Status = 5 // RequestStatusError encountered error
)

func (rs Status) ToString() string {
	switch rs {
	case StatusNone:
		return "None"
	case StatusInProgress:
		return "InProgress"
	case StatusCompleted:
		return "Completed"
	case StatusCanceled:
		return "Canceled"
	case StatusError:
		return "Error"
	}
	return ""
}
