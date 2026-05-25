package models

import "slices"

type MWUser struct {
	ID          int
	Name        string
	Groups      []string
	BlockID     *int
	BlockedByID *int
	BlockedBy   *string
	BlockExpiry *string
}

func (u MWUser) IsSysop() bool {
	return slices.Contains(u.Groups, "sysop")
}

func (u MWUser) IsBlocked() bool {
	return u.BlockID != nil || u.BlockedByID != nil || u.BlockedBy != nil || u.BlockExpiry != nil
}
