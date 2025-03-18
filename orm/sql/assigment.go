package sql

type Assignment struct {
	Col string
	Val any
}

func (Assignment) Assign() {}

func Assign(col string, val any) Assignment {
	return Assignment{
		Col: col,
		Val: val,
	}
}
