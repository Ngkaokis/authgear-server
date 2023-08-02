package workflowconfig

import (
	"context"
	"fmt"
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentLoginFlowStepAuthenticate{})
}

var IntentLoginFlowStepAuthenticateSchema = validation.NewSimpleSchema(`{}`)

type IntentLoginFlowStepAuthenticate struct {
	LoginFlow   string        `json:"login_flow,omitempty"`
	JSONPointer jsonpointer.T `json:"json_pointer,omitempty"`
	StepID      string        `json:"step_id,omitempty"`
	UserID      string        `json:"user_id,omitempty"`
}

var _ WorkflowStep = &IntentLoginFlowStepAuthenticate{}

func (i *IntentLoginFlowStepAuthenticate) GetID() string {
	return i.StepID
}

func (i *IntentLoginFlowStepAuthenticate) GetJSONPointer() jsonpointer.T {
	return i.JSONPointer
}

var _ workflow.Intent = &IntentLoginFlowStepAuthenticate{}

func NewIntentLoginFlowStepAuthenticate(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, i *IntentLoginFlowStepAuthenticate) (*IntentLoginFlowStepAuthenticate, error) {
	// OutputData will include usable authenticators.
	// If it returns error, there is no usable authenticators.
	// This intent cannot proceed if there is no usable authenticators.
	// Therefore, we prevent from adding this intent to the workflow if such case happens.
	_, err := i.OutputData(ctx, deps, workflows)
	if err != nil {
		return nil, err
	}

	return i, nil
}

func (*IntentLoginFlowStepAuthenticate) Kind() string {
	return "workflowconfig.IntentLoginFlowStepAuthenticate"
}

func (*IntentLoginFlowStepAuthenticate) JSONSchema() *validation.SimpleSchema {
	return IntentLoginFlowStepAuthenticateSchema
}

func (*IntentLoginFlowStepAuthenticate) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	// Let the input to select which authentication method to use.
	if len(workflows.Nearest.Nodes) == 0 {
		return []workflow.Input{
			&InputTakeAuthenticationMethod{},
		}, nil
	}

	lastNode := workflows.Nearest.Nodes[len(workflows.Nearest.Nodes)-1]
	if lastNode.Type == workflow.NodeTypeSimple {
		switch lastNode.Simple.(type) {
		case *NodeDoUseAuthenticator:
			// Handle nested steps.
			return nil, nil
		}
	}

	return nil, workflow.ErrEOF
}

func (i *IntentLoginFlowStepAuthenticate) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	if len(workflows.Nearest.Nodes) == 0 {
		var inputTakeAuthenticationMethod inputTakeAuthenticationMethod
		if workflow.AsInput(input, &inputTakeAuthenticationMethod) &&
			inputTakeAuthenticationMethod.GetJSONPointer().String() == i.JSONPointer.String() {
			authentication := inputTakeAuthenticationMethod.GetAuthenticationMethod()
			_, err := i.checkAuthenticationMethod(deps, authentication)
			if err != nil {
				return nil, err
			}

			switch authentication {
			case config.WorkflowAuthenticationMethodPrimaryPassword:
				fallthrough
			case config.WorkflowAuthenticationMethodSecondaryPassword:
				return workflow.NewNodeSimple(&NodeUseAuthenticatorPassword{
					UserID:         i.UserID,
					Authentication: authentication,
				}), nil
			case config.WorkflowAuthenticationMethodPrimaryPasskey:
				// FIXME(workflow): authenticate with passkey
			case config.WorkflowAuthenticationMethodPrimaryOOBOTPEmail:
				fallthrough
			case config.WorkflowAuthenticationMethodSecondaryOOBOTPEmail:
				fallthrough
			case config.WorkflowAuthenticationMethodPrimaryOOBOTPSMS:
				fallthrough
			case config.WorkflowAuthenticationMethodSecondaryOOBOTPSMS:
				// FIXME(workflow): authenticate with oob otp
			case config.WorkflowAuthenticationMethodSecondaryTOTP:
				// FIXME(workflow): authenticate with totp
			case config.WorkflowAuthenticationMethodRecoveryCode:
				// FIXME(workflow): authenticate with recovery code
			case config.WorkflowAuthenticationMethodDeviceToken:
				// FIXME(workflow): authenticate with device token
			}
		}

		return nil, workflow.ErrIncompatibleInput
	}

	current, err := loginFlowCurrent(deps, i.LoginFlow, i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	lastNode := workflows.Nearest.Nodes[len(workflows.Nearest.Nodes)-1]
	if lastNode.Type == workflow.NodeTypeSimple {
		switch lastNode.Simple.(type) {
		case *NodeDoUseAuthenticator:
			identification := i.authenticationMethod(workflows.Nearest)
			return workflow.NewSubWorkflow(&IntentLoginFlowSteps{
				LoginFlow:   i.LoginFlow,
				JSONPointer: i.jsonPointer(step, identification),
			}), nil
		}
	}

	return nil, workflow.ErrIncompatibleInput
}

func (*IntentLoginFlowStepAuthenticate) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Effect, error) {
	return nil, nil
}

func (i *IntentLoginFlowStepAuthenticate) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	current, err := loginFlowCurrent(deps, i.LoginFlow, i.JSONPointer)
	if err != nil {
		return nil, err
	}
	step := i.step(current)

	allAllowed := i.getAllAllowed(step)
	allUsable, err := getAuthenticationMethodsOfUser(deps, i.UserID, allAllowed)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"json_pointer":           i.JSONPointer.String(),
		"authentication_methods": allUsable,
	}, nil
}

func (i *IntentLoginFlowStepAuthenticate) checkAuthenticationMethod(deps *workflow.Dependencies, am config.WorkflowAuthenticationMethod) (idx int, err error) {
	idx = -1

	current, err := loginFlowCurrent(deps, i.LoginFlow, i.JSONPointer)
	if err != nil {
		return
	}
	step := i.step(current)

	allAllowed := i.getAllAllowed(step)
	allUsable, err := getAuthenticationMethodsOfUser(deps, i.UserID, allAllowed)
	if err != nil {
		return
	}

	for i, usable := range allUsable {
		if am == usable {
			idx = i
		}
	}

	if idx >= 0 {
		return
	}

	err = InvalidAuthenticationMethod.NewWithInfo("invalid authentication method", apierrors.Details{
		"expected": allUsable,
		"actual":   am,
	})
	return
}

func (*IntentLoginFlowStepAuthenticate) getAllAllowed(step *config.WorkflowLoginFlowStep) []config.WorkflowAuthenticationMethod {
	// Make empty slice.
	allAllowed := []config.WorkflowAuthenticationMethod{}

	for _, branch := range step.OneOf {
		branch := branch
		allAllowed = append(allAllowed, branch.Authentication)
	}

	return allAllowed
}

func (*IntentLoginFlowStepAuthenticate) step(o config.WorkflowObject) *config.WorkflowLoginFlowStep {
	step, ok := o.(*config.WorkflowLoginFlowStep)
	if !ok {
		panic(fmt.Errorf("workflow: workflow object is %T", o))
	}

	return step
}

func (*IntentLoginFlowStepAuthenticate) authenticationMethod(w *workflow.Workflow) config.WorkflowAuthenticationMethod {
	if len(w.Nodes) == 0 {
		panic(fmt.Errorf("workflow: authentication method not yet selected"))
	}

	switch n := w.Nodes[0].Simple.(type) {
	case *NodeUseAuthenticatorPassword:
		return n.Authentication
	default:
		panic(fmt.Errorf("workflow: unexpected node: %T", w.Nodes[0].Simple))
	}
}

func (i *IntentLoginFlowStepAuthenticate) jsonPointer(step *config.WorkflowLoginFlowStep, am config.WorkflowAuthenticationMethod) jsonpointer.T {
	for idx, branch := range step.OneOf {
		branch := branch
		if branch.Authentication == am {
			return JSONPointerForOneOf(i.JSONPointer, idx)
		}
	}

	panic(fmt.Errorf("workflow: selected authentication method is not allowed"))
}
