package network

import (
	"fmt"
)

// additional functions to complement network.pb.go

func (s *StaticId) CompactId() string {
	return fmt.Sprintf("%x-%d", s.GetHash(), s.GetLength())
}
