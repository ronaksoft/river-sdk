package supernumerary

import (
	ronak "git.ronaksoftware.com/ronak/toolbox"
)

type Supernumerary struct {
	Metrics *ronak.Prometheus
}

func NewSupernumerary() *Supernumerary {
	return &Supernumerary{}
}

func (s *Supernumerary) Init(from, to int) {

}
