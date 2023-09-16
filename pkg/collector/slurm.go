package collector

import "fmt"

type SlurmProcInfo struct {
	SlurmJobID   string `json:"slurmJobID"`
	SlurmStepID  string `json:"slurmStepID"`
	SlurmUser    string `json:"slurmUser"`
	SlurmAccount string `json:"slurmAccount"`
	SlurmJobName string `json:"slurmJobName"`
}

const (
	// Slurm Process Env Key
	SLURM_ENV_JOBID   = "SLURM_JOBID"
	SLURM_ENV_STEP_ID = "SLURM_STEP_ID"
	SLURM_ENV_USER    = "SLURM_JOB_USER"
	SLURM_ENV_ACCOUNT = "SLURM_JOB_ACCOUNT"
	SLURM_ENV_JOBNAME = "SLURM_JOB_NAME"
)

var (
	SlurmProcLabels = []string{
		"gpu", "pid", "procName", "user",
		"slurmJobID", "slurmStepID", "slurmUser", "slurmAccount", "slurmJobName",
	}
	getSlurmProcessStatLabelValues = func(ps ProcessStat) []string {
		// todo: json unmarshall
		return []string{
			fmt.Sprintf("%d", ps.GPUIndex),
			fmt.Sprintf("%d", ps.Pid),
			ps.ProcName,
			ps.User,
			ps.SlurmJobID,
			ps.SlurmStepID,
			ps.SlurmUser,
			ps.SlurmAccount,
			ps.SlurmJobName,
		}
	}
)
