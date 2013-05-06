package network

import (
	"fmt"
)

// additional functions not provided by network.pb.go

func (s *StaticId) CompactId() string {
	return fmt.Sprintf("%x-%d", s.GetHash(), s.GetLength())
}
