package models

type AnalysisResult struct {
	RawOutput string
	Errors    []string
}

type CodeMetrics struct {
	CyclomaticComplexity int
	DuplicateCode        int
	TestCoverage         float64
	FunctionCount        int
	LongFunctions        int
	DeadCodeCount        int
}

type Report struct {
	Language    string
	CodeMetrics CodeMetrics
	Suggestions []string
}
