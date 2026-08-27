package rbac

type EffectiveRBAC struct {
	IsSuperuser bool
	Permissions map[string]bool
	RoleNames   []string
}

func (e *EffectiveRBAC) Has(p string) bool {
	if e == nil {
		return false
	}
	if e.IsSuperuser {
		return true
	}
	return e.Permissions[p]
}
