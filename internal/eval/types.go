package eval

type EvalCase struct {
	Task           string   `json:"task"`
	Requirements   string   `json:"requirements,omitempty"`
	Constraints    string   `json:"constraints,omitempty"`
	ExpectedTopics []string `json:"expectedTopics,omitempty"`
	MinScore       float64  `json:"minScore,omitempty"`
}

type EvalScore struct {
	Name    string  `json:"name"`
	Score   float64 `json:"score"`
	Max     float64 `json:"max"`
	Pass    bool    `json:"pass"`
	Details string  `json:"details,omitempty"`
}

type EvalReport struct {
	Task         string      `json:"task"`
	Model        string      `json:"model"`
	Scores       []EvalScore `json:"scores"`
	OverallScore float64     `json:"overallScore"`
	Pass         bool        `json:"pass"`
	Summary      string      `json:"summary"`
	Markdown     string      `json:"markdown"`
}

type Scorer interface {
	Name() string
	Max() float64
	Score(text string, evalCase *EvalCase) (float64, error)
}

type EvalOptions struct {
	Judge string
}
