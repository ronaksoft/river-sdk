package uiexec

var (
	funcChan = make(chan func(), 100)
)

func init() {
	go func() {
		for fn := range funcChan {
			fn()
		}
	}()

}

// Exec pass given function to UIExecutor buffered channel
func Exec(fn func()) {
	funcChan <- fn
}

