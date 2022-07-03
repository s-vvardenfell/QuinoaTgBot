package conditions

type Conditions struct {
	Type      string
	Genres    []string
	StartYear string
	EndYear   string
	Countries []string
	Keyword   string
	Check     ConditionsCheck
}

// используется для "опроса"
type ConditionsCheck struct {
	Type      bool
	Genres    bool
	StartYear bool
	EndYear   bool
	Countries bool
	Keyword   bool
}
