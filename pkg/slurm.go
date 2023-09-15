package collector

type SlurmProcInfo struct {
	JobID        uint   `json:"slurmJobID"`
	StepID       string `json:"slurmStepID"`
	SlurmUser    string `json:"slurmUser"`
	SlurmAccount string `json:"slurmAccount"`
}

const (
	LabelJobID        = "slurmJobID"
	LabelStepID       = "slurmStepID"
	LabelSlurmUser    = "slurmUser"
	LabelSlurmAccount = "slurmAccount"
)

var (
	SlurmProcLabels = []string{LabelGPU, LabelHostName, LabelPID, LabelProcName, LabelUser, LabelJobID, LabelStepID, LabelSlurmUser, LabelSlurmAccount}
)
