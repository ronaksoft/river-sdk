package supernumerary

import (
	"git.ronaksoftware.com/ronak/riversdk/loadtester/scenario"
	"git.ronaksoftware.com/ronak/riversdk/loadtester/shared"
)

func playCreateAuthKey(act shared.Acter, isFinal bool) {
	s := scenario.NewCreateAuthKey(isFinal)
	scenario.Play(act, s)
	s.Wait(act)
}
