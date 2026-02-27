package agent

import "github.com/google/uuid"

func NewWorkflowID() string {
    return uuid.NewString()
}
