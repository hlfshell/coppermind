package agents

type Agent struct {
	ID       string `json:"id,omitempty" db:"id"`
	Name     string `json:"name,omitempty" db:"name"`
	Identity string `json:"identity,omitempty" db:"identity"`
}
