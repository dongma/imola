package sql

// Aggregate 代表了聚合函数，AVG("Age"),SUM("Age"),Count("Age"),MAX("Age"),MIN("Age")
type Aggregate struct {
	Fn    string
	Arg   string
	Alias string
}

func (a Aggregate) Selectable() {
}

func (a Aggregate) As(alias string) Aggregate {
	return Aggregate{
		Fn:    a.Fn,
		Arg:   a.Arg,
		Alias: alias,
	}
}

func Avg(col string) Aggregate {
	return Aggregate{
		Fn:  "AVG",
		Arg: col,
	}
}

func Sum(col string) Aggregate {
	return Aggregate{
		Fn:  "SUM",
		Arg: col,
	}
}

func Count(col string) Aggregate {
	return Aggregate{
		Fn:  "COUNT",
		Arg: col,
	}
}

func Max(col string) Aggregate {
	return Aggregate{
		Fn:  "MAX",
		Arg: col,
	}
}

func Min(col string) Aggregate {
	return Aggregate{
		Fn:  "MIN",
		Arg: col,
	}
}
